// Copyright (c) 2022 Alibaba Group Holding Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

#include "extensions/oauth/plugin.h"

#include <algorithm>
#include <array>
#include <chrono>
#include <cstdint>
#include <limits>
#include <memory>
#include <optional>
#include <random>
#include <string>
#include <string_view>
#include <unordered_set>
#include <utility>

#include "absl/strings/match.h"
#include "absl/strings/str_cat.h"
#include "absl/strings/str_format.h"
#include "absl/strings/str_join.h"
#include "absl/strings/str_split.h"
#include "common/common_util.h"
#include "common/http_util.h"
#include "common/json_util.h"
#include "openssl/hmac.h"

using ::nlohmann::json;
using ::Wasm::Common::JsonArrayIterate;
using ::Wasm::Common::JsonGetField;
using ::Wasm::Common::JsonObjectIterate;
using ::Wasm::Common::JsonValueAs;

#ifdef NULL_PLUGIN

namespace proxy_wasm {
namespace null_plugin {
namespace oauth {

PROXY_WASM_NULL_PLUGIN_REGISTRY

#endif
namespace {
constexpr absl::string_view TokenResponseTemplate = R"(
{
  "token_type": "bearer",
  "access_token": "%s",
  "expires_in": %u
})";
const std::string& DefaultAudience = "default";
const std::string& TypeHeader = "application/at+jwt";
const std::string& BearerPrefix = "Bearer ";
const std::string& ClientCredentialsGrant = "client_credentials";
constexpr uint32_t MaximumUriLength = 256;
constexpr std::string_view kRcDetailOAuthPrefix = "oauth_access_denied";
constexpr std::string_view kJwtAlg = "HS256";

bool base64UrlEncode(std::string_view input, std::string* output) {
  if (output == nullptr) {
    return false;
  }
  static constexpr char kEncodeTable[] =
      "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_";

  output->clear();
  output->reserve((input.size() * 4 + 2) / 3);
  uint32_t value = 0;
  int valb = -6;
  for (const uint8_t ch : input) {
    value = (value << 8) | ch;
    valb += 8;
    while (valb >= 0) {
      output->push_back(kEncodeTable[(value >> valb) & 0x3f]);
      valb -= 6;
    }
  }
  if (valb > -6) {
    output->push_back(kEncodeTable[((value << 8) >> (valb + 8)) & 0x3f]);
  }
  return true;
}

bool base64UrlDecode(std::string_view input, std::string* output) {
  if (output == nullptr) {
    return false;
  }
  if (input.size() % 4 == 1) {
    return false;
  }
  output->clear();
  output->reserve((input.size() * 3) / 4);
  uint32_t value = 0;
  int valb = -8;
  for (const uint8_t ch : input) {
    if (ch == '=') {
      break;
    }
    int8_t decoded = -1;
    if (ch >= 'A' && ch <= 'Z') {
      decoded = static_cast<int8_t>(ch - 'A');
    } else if (ch >= 'a' && ch <= 'z') {
      decoded = static_cast<int8_t>(ch - 'a' + 26);
    } else if (ch >= '0' && ch <= '9') {
      decoded = static_cast<int8_t>(ch - '0' + 52);
    } else if (ch == '-') {
      decoded = 62;
    } else if (ch == '_') {
      decoded = 63;
    }
    if (decoded < 0) {
      return false;
    }
    value = (value << 6) | static_cast<uint32_t>(decoded);
    valb += 6;
    if (valb >= 0) {
      output->push_back(static_cast<char>((value >> valb) & 0xff));
      valb -= 8;
    }
  }
  return true;
}

bool getJsonStringField(const json& src, std::string_view field,
                        std::string* value) {
  if (value == nullptr) {
    return false;
  }
  auto field_json = src.find(std::string(field));
  if (field_json == src.end() || !field_json->is_string()) {
    return false;
  }
  *value = field_json->get<std::string>();
  return true;
}

bool getJsonIntegerField(const json& src, std::string_view field,
                         uint64_t* value) {
  if (value == nullptr) {
    return false;
  }
  auto field_json = src.find(std::string(field));
  if (field_json == src.end()) {
    return false;
  }
  if (field_json->is_number_unsigned()) {
    *value = field_json->get<uint64_t>();
    return true;
  }
  if (!field_json->is_number_integer()) {
    return false;
  }
  auto int_value = field_json->get<int64_t>();
  if (int_value < 0) {
    return false;
  }
  *value = static_cast<uint64_t>(int_value);
  return true;
}

bool constantTimeEquals(std::string_view lhs, std::string_view rhs) {
  if (lhs.size() != rhs.size()) {
    return false;
  }
  uint8_t diff = 0;
  for (size_t i = 0; i < lhs.size(); ++i) {
    diff |= static_cast<uint8_t>(lhs[i] ^ rhs[i]);
  }
  return diff == 0;
}

bool signHs256(std::string_view key, std::string_view input,
               std::string* signature) {
  if (signature == nullptr) {
    return false;
  }
  std::array<uint8_t, EVP_MAX_MD_SIZE> hmac{};
  unsigned int length = 0;
  auto* result = HMAC(EVP_sha256(), key.data(), static_cast<int>(key.size()),
                      reinterpret_cast<const uint8_t*>(input.data()),
                      input.size(), hmac.data(), &length);
  if (result == nullptr || length == 0) {
    return false;
  }
  signature->assign(reinterpret_cast<const char*>(hmac.data()), length);
  return true;
}

bool addWithoutOverflow(uint64_t lhs, uint64_t rhs, uint64_t* value) {
  if (value == nullptr) {
    return false;
  }
  if (lhs > std::numeric_limits<uint64_t>::max() - rhs) {
    *value = std::numeric_limits<uint64_t>::max();
    return true;
  }
  *value = lhs + rhs;
  return true;
}

bool splitToken(std::string_view token, std::array<std::string_view, 3>* parts) {
  if (parts == nullptr) {
    return false;
  }
  auto first_dot = token.find('.');
  if (first_dot == std::string::npos) {
    return false;
  }
  auto second_dot = token.find('.', first_dot + 1);
  if (second_dot == std::string::npos || token.find('.', second_dot + 1) !=
                                         std::string::npos) {
    return false;
  }
  (*parts)[0] = token.substr(0, first_dot);
  (*parts)[1] = token.substr(first_dot + 1, second_dot - first_dot - 1);
  (*parts)[2] = token.substr(second_dot + 1);
  return !(*parts)[0].empty() && !(*parts)[1].empty() && !(*parts)[2].empty();
}

bool parseJsonSegment(std::string_view encoded, json* out) {
  if (out == nullptr) {
    return false;
  }
  std::string decoded;
  if (!base64UrlDecode(encoded, &decoded)) {
    return false;
  }
  *out = json::parse(decoded, nullptr, false);
  return out->is_object();
}

std::string toHex(uint8_t value) {
  static constexpr char kHex[] = "0123456789abcdef";
  std::string out(2, '0');
  out[0] = kHex[(value >> 4) & 0x0f];
  out[1] = kHex[value & 0x0f];
  return out;
}

std::string generateUuidV4(std::mt19937* generator) {
  if (generator == nullptr) {
    return {};
  }
  std::uniform_int_distribution<int> dist(0, 255);
  std::array<uint8_t, 16> bytes{};
  for (auto& byte : bytes) {
    byte = static_cast<uint8_t>(dist(*generator));
  }
  // RFC 4122 variant 1 + version 4.
  bytes[6] = (bytes[6] & 0x0f) | 0x40;
  bytes[8] = (bytes[8] & 0x3f) | 0x80;

  std::string uuid;
  uuid.reserve(36);
  for (size_t i = 0; i < bytes.size(); ++i) {
    if (i == 4 || i == 6 || i == 8 || i == 10) {
      uuid.push_back('-');
    }
    uuid.append(toHex(bytes[i]));
  }
  return uuid;
}

bool buildToken(std::string_view secret, const json& payload, std::string* token) {
  if (token == nullptr) {
    return false;
  }
  const json header = {{"alg", kJwtAlg}, {"typ", TypeHeader}};
  std::string header_segment;
  std::string payload_segment;
  if (!base64UrlEncode(header.dump(), &header_segment) ||
      !base64UrlEncode(payload.dump(), &payload_segment)) {
    return false;
  }
  std::string signing_input = absl::StrCat(header_segment, ".", payload_segment);
  std::string signature;
  if (!signHs256(secret, signing_input, &signature)) {
    return false;
  }
  std::string signature_segment;
  if (!base64UrlEncode(signature, &signature_segment)) {
    return false;
  }
  *token = absl::StrCat(signing_input, ".", signature_segment);
  return true;
}

std::string generateRcDetails(std::string_view error_msg) {
  // Replace space with underscore since RCDetails may be written to access log.
  // Some log processors assume each log segment is separated by whitespace.
  return absl::StrCat(kRcDetailOAuthPrefix, "{",
                      absl::StrJoin(absl::StrSplit(error_msg, ' '), "_"), "}");
}
}  // namespace
static RegisterContextFactory register_OAuth(CONTEXT_FACTORY(PluginContext),
                                             ROOT_FACTORY(PluginRootContext));

#define JSON_FIND_FIELD(dict, field)               \
  auto dict##_##field##_json = dict.find(#field);  \
  if (dict##_##field##_json == dict.end()) {       \
    LOG_WARN("can't find '" #field "' in " #dict); \
    return false;                                  \
  }

#define JSON_VALUE_AS(type, src, dst, err_msg)                      \
  auto dst##_v = JsonValueAs<type>(src);                            \
  if (dst##_v.second != Wasm::Common::JsonParserResultDetail::OK || \
      !dst##_v.first) {                                             \
    LOG_WARN(#err_msg);                                             \
    return false;                                                   \
  }                                                                 \
  auto& dst = dst##_v.first.value();

#define JSON_FIELD_VALUE_AS(type, dict, field)                       \
  JSON_VALUE_AS(type, dict##_##field##_json.value(), dict##_##field, \
                "'" #field "' field in " #dict "convert to " #type " failed")

bool PluginRootContext::generateToken(const OAuthConfigRule& rule,
                                      const std::string& route_name,
                                      const absl::string_view& raw_params,
                                      std::string* token,
                                      std::string* err_msg) {
  auto params = Wasm::Common::Http::parseParameters(raw_params, 0, true);
  auto it = params.find("grant_type");
  if (it == params.end()) {
    *err_msg = "grant_type is missing";
    return false;
  }
  if (it->second != ClientCredentialsGrant) {
    *err_msg = absl::StrFormat("grant_type:%s is not support", it->second);
    return false;
  }
  it = params.find("client_id");
  if (it == params.end()) {
    *err_msg = "client_id is missing";
    return false;
  }
  auto c_it = rule.consumers.find(it->second);
  if (c_it == rule.consumers.end()) {
    *err_msg = "invalid client_id or client_secret";
    return false;
  }
  const auto& consumer = c_it->second;
  it = params.find("client_secret");
  if (it == params.end()) {
    *err_msg = "client_secret is missing";
    return false;
  }
  if (it->second != consumer.client_secret) {
    *err_msg = "invalid client_id or client_secret";
    return false;
  }
  std::random_device rd;
  auto seed_data = std::array<int, std::mt19937::state_size>{};
  std::generate(std::begin(seed_data), std::end(seed_data), std::ref(rd));
  std::seed_seq seq(std::begin(seed_data), std::end(seed_data));
  std::mt19937 generator(seq);
  const auto now = std::chrono::system_clock::now();
  const uint64_t now_seconds =
      std::chrono::duration_cast<std::chrono::seconds>(now.time_since_epoch())
          .count();
  uint64_t exp_seconds = 0;
  if (!addWithoutOverflow(now_seconds, rule.token_ttl, &exp_seconds)) {
    *err_msg = "jwt sign failed";
    return false;
  }

  json payload = {{"iss", rule.issuer},
                  {"sub", consumer.name},
                  {"iat", now_seconds},
                  {"exp", exp_seconds},
                  {"client_id", consumer.client_id},
                  {"jti", generateUuidV4(&generator)}};
  payload["aud"] = rule.global_credentials ? json(DefaultAudience)
                                           : json(route_name);

  it = params.find("scope");
  if (it != params.end()) {
    payload["scope"] = it->second;
  }

  if (!buildToken(consumer.client_secret, payload, token)) {
    *err_msg = "jwt sign failed";
    return false;
  }
  return true;
}

bool PluginRootContext::parsePluginConfig(const json& conf,
                                          OAuthConfigRule& rule) {
  std::unordered_set<std::string> name_set;
  if (!JsonArrayIterate(conf, "consumers", [&](const json& consumer) -> bool {
        Consumer c;
        JSON_FIND_FIELD(consumer, name);
        JSON_FIELD_VALUE_AS(std::string, consumer, name);
        if (name_set.count(consumer_name) != 0) {
          LOG_WARN("consumer already exists: " + consumer_name);
          return false;
        }
        c.name = consumer_name;
        JSON_FIND_FIELD(consumer, client_id);
        JSON_FIELD_VALUE_AS(std::string, consumer, client_id);
        c.client_id = consumer_client_id;
        if (rule.consumers.find(c.client_id) != rule.consumers.end()) {
          LOG_WARN("consumer client_id already exists: " + c.client_id);
          return false;
        }
        JSON_FIND_FIELD(consumer, client_secret);
        JSON_FIELD_VALUE_AS(std::string, consumer, client_secret);
        c.client_secret = consumer_client_secret;
        rule.consumers.emplace(c.client_id, std::move(c));
        name_set.insert(consumer_name);
        return true;
      })) {
    LOG_WARN("failed to parse configuration for consumers.");
    return false;
  }
  // if (rule.consumers.empty()) {
  //   LOG_INFO("at least one consumer has to be configured for a rule.");
  //   return false;
  // }
  auto conf_issuer_json = conf.find("issuer");
  if (conf_issuer_json != conf.end()) {
    JSON_FIELD_VALUE_AS(std::string, conf, issuer);
    rule.issuer = conf_issuer;
  }
  auto conf_auth_header_json = conf.find("auth_header");
  if (conf_auth_header_json != conf.end()) {
    JSON_FIELD_VALUE_AS(std::string, conf, auth_header);
    rule.auth_header_name = conf_auth_header;
  }
  auto conf_auth_path_json = conf.find("auth_path");
  if (conf_auth_path_json != conf.end()) {
    JSON_FIELD_VALUE_AS(std::string, conf, auth_path);
    if (conf_auth_path.empty()) {
      conf_auth_path = "/";
    } else if (conf_auth_path[0] != '/') {
      conf_auth_path = absl::StrCat("/", conf_auth_path);
    }
    rule.auth_path = conf_auth_path;
  }
  auto conf_global_credentials_json = conf.find("global_credentials");
  if (conf_global_credentials_json != conf.end()) {
    JSON_FIELD_VALUE_AS(bool, conf, global_credentials);
    rule.global_credentials = conf_global_credentials;
  }
  auto conf_token_ttl_json = conf.find("token_ttl");
  if (conf_token_ttl_json != conf.end()) {
    JSON_FIELD_VALUE_AS(uint64_t, conf, token_ttl);
    rule.token_ttl = conf_token_ttl;
  }
  auto conf_keep_token_json = conf.find("keep_token");
  if (conf_keep_token_json != conf.end()) {
    JSON_FIELD_VALUE_AS(bool, conf, keep_token);
    rule.keep_token = conf_keep_token;
  }
  auto conf_clock_skew_seconds_json = conf.find("clock_skew_seconds");
  if (conf_clock_skew_seconds_json != conf.end()) {
    JSON_FIELD_VALUE_AS(uint64_t, conf, clock_skew_seconds);
    rule.clock_skew = conf_clock_skew_seconds;
  }
  return true;
}

bool PluginRootContext::checkPlugin(
    const OAuthConfigRule& rule,
    const std::optional<std::unordered_set<std::string>>& allow_set,
    const std::string& route_name) {
  auto auth_header = getRequestHeader(rule.auth_header_name)->toString();
  bool verified = false;
  std::string token_str;
  {
    size_t pos = 0;
    if (auth_header.empty()) {
      LOG_DEBUG("auth header is empty");
      goto failed;
    }
    pos = auth_header.find(BearerPrefix);
    if (pos == std::string::npos) {
      LOG_DEBUG("auth header is not a bearer token");
      goto failed;
    }
    auto start = pos + BearerPrefix.size();
    token_str =
        std::string{auth_header.c_str() + start, auth_header.size() - start};
    std::array<std::string_view, 3> token_parts;
    if (!splitToken(token_str, &token_parts)) {
      LOG_DEBUG(absl::StrFormat("invalid token format, token:%s", token_str));
      goto failed;
    }

    json header_json;
    if (!parseJsonSegment(token_parts[0], &header_json)) {
      LOG_DEBUG(absl::StrFormat("invalid token header, token:%s", token_str));
      goto failed;
    }

    std::string alg;
    if (!getJsonStringField(header_json, "alg", &alg) || alg != kJwtAlg) {
      LOG_DEBUG(absl::StrFormat("invalid token alg, token:%s", token_str));
      goto failed;
    }
    std::string typ;
    if (!getJsonStringField(header_json, "typ", &typ) || typ != TypeHeader) {
      LOG_DEBUG(absl::StrFormat("invalid token typ, token:%s", token_str));
      goto failed;
    }

    json payload_json;
    if (!parseJsonSegment(token_parts[1], &payload_json)) {
      LOG_DEBUG(absl::StrFormat("invalid token payload, token:%s", token_str));
      goto failed;
    }

    std::string client_id;
    if (!getJsonStringField(payload_json, "client_id", &client_id)) {
      LOG_DEBUG("claim is missing or invalid: client_id");
      goto failed;
    }
    auto it = rule.consumers.find(client_id);
    if (it == rule.consumers.end()) {
      LOG_DEBUG(absl::StrFormat("client_id not found:%s", client_id));
      goto failed;
    }
    const auto& consumer = it->second;

    const std::string signing_input =
        absl::StrCat(token_parts[0], ".", token_parts[1]);
    std::string expected_signature;
    if (!signHs256(consumer.client_secret, signing_input, &expected_signature)) {
      LOG_INFO(absl::StrFormat("token verify failed, token:%s, reason:%s",
                               token_str, "sign failed"));
      goto failed;
    }
    std::string expected_signature_segment;
    if (!base64UrlEncode(expected_signature, &expected_signature_segment)) {
      LOG_INFO(absl::StrFormat("token verify failed, token:%s, reason:%s",
                               token_str, "base64 encode failed"));
      goto failed;
    }
    if (!constantTimeEquals(expected_signature_segment, token_parts[2])) {
      LOG_INFO(absl::StrFormat("token verify failed, token:%s, reason:%s",
                               token_str, "signature invalid"));
      goto failed;
    }

    std::string issuer;
    std::string subject;
    std::string audience;
    uint64_t expire_at = 0;
    uint64_t issued_at = 0;
    if (!getJsonStringField(payload_json, "iss", &issuer) ||
        !getJsonStringField(payload_json, "sub", &subject) ||
        !getJsonStringField(payload_json, "aud", &audience) ||
        !getJsonIntegerField(payload_json, "exp", &expire_at) ||
        !getJsonIntegerField(payload_json, "iat", &issued_at)) {
      LOG_DEBUG(absl::StrFormat("token required claim is invalid, token:%s",
                                token_str));
      goto failed;
    }
    if (issuer != rule.issuer || subject != consumer.name) {
      LOG_INFO(absl::StrFormat("token verify failed, token:%s, reason:%s",
                               token_str, "iss/sub mismatch"));
      goto failed;
    }

    const uint64_t now = std::chrono::duration_cast<std::chrono::seconds>(
                             std::chrono::system_clock::now().time_since_epoch())
                             .count();
    uint64_t now_with_leeway = 0;
    uint64_t expire_with_leeway = 0;
    addWithoutOverflow(now, rule.clock_skew, &now_with_leeway);
    addWithoutOverflow(expire_at, rule.clock_skew, &expire_with_leeway);
    if (issued_at > now_with_leeway || now > expire_with_leeway) {
      LOG_INFO(absl::StrFormat("token verify failed, token:%s, reason:%s",
                               token_str, "time constraint"));
      goto failed;
    }

    verified = true;
    if (allow_set &&
        allow_set.value().find(consumer.name) == allow_set.value().end()) {
      LOG_DEBUG(absl::StrFormat("consumer:%s is not in route's:%s allow_set",
                                consumer.name, route_name));
      goto failed;
    }
    if (!rule.global_credentials) {
      if (audience != route_name) {
        LOG_DEBUG(absl::StrFormat("audience:%s not match this route:%s",
                                  audience, route_name));
        goto failed;
      }
    }
    if (!rule.keep_token) {
      removeRequestHeader(rule.auth_header_name);
    }
    addRequestHeader("X-Mse-Consumer", consumer.name);
    return true;
  }
failed:
  if (!verified) {
    auto authn_value = absl::StrCat(
        "Bearer realm=\"",
        Wasm::Common::Http::buildOriginalUri(MaximumUriLength), "\"");
    sendLocalResponse(401, kRcDetailOAuthPrefix, "Invalid Jwt token",
                      {{"WWW-Authenticate", authn_value}});
  } else {
    sendLocalResponse(403, kRcDetailOAuthPrefix, "Access Denied", {});
  }
  return false;
}

bool PluginRootContext::onConfigure(size_t size) {
  // Parse configuration JSON string.
  if (size > 0 && !configure(size)) {
    LOG_WARN("configuration has errors initialization will not continue.");
    setInvalidConfig();
    return false;
  }
  return true;
}

bool PluginRootContext::configure(size_t configuration_size) {
  auto configuration_data = getBufferBytes(WasmBufferType::PluginConfiguration,
                                           0, configuration_size);
  // Parse configuration JSON string.
  auto result = ::Wasm::Common::JsonParse(configuration_data->view());
  if (!result) {
    LOG_WARN(absl::StrCat("cannot parse plugin configuration JSON string: ",
                          configuration_data->view()));
    return false;
  }
  if (!parseAuthRuleConfig(result.value())) {
    LOG_WARN(absl::StrCat("cannot parse plugin configuration JSON string: ",
                          configuration_data->view()));
    return false;
  }
  return true;
}

FilterHeadersStatus PluginContext::onRequestHeaders(uint32_t, bool) {
  auto* rootCtx = rootContext();
  auto config = rootCtx->getMatchAuthConfig();
  if (!config.first) {
    return FilterHeadersStatus::Continue;
  }
  config_ = config.first;
  getValue({"route_name"}, &route_name_);
  auto path = getRequestHeader(Wasm::Common::Http::Header::Path)->toString();
  auto params_pos = path.find('?');
  size_t uri_end;
  if (params_pos == std::string::npos) {
    uri_end = path.size();
  } else {
    uri_end = params_pos;
  }
  // Authorize request
  if (absl::EndsWith({path.c_str(), uri_end},
                     config_.value().get().auth_path)) {
    std::string err_msg, token;
    auto method =
        getRequestHeader(Wasm::Common::Http::Header::Method)->toString();
    if (method == "GET") {
      if (params_pos == std::string::npos) {
        err_msg = "Authorize parameters are missing";
        goto done;
      }
      params_pos++;
      rootCtx->generateToken(
          config_.value(), route_name_,
          {path.c_str() + params_pos, path.size() - params_pos}, &token,
          &err_msg);
      goto done;
    }
    if (method == "POST") {
      auto content_type =
          getRequestHeader(Wasm::Common::Http::Header::ContentType)->toString();
      if (!absl::StrContains(absl::AsciiStrToLower(content_type),
                             "application/x-www-form-urlencoded")) {
        err_msg = "Invalid content-type";
        goto done;
      }
      check_body_params_ = true;
    }
  done:
    if (!err_msg.empty()) {
      sendLocalResponse(400, generateRcDetails(err_msg), err_msg, {});
      return FilterHeadersStatus::StopIteration;
    }
    if (!token.empty()) {
      sendLocalResponse(200, "",
                        absl::StrFormat(TokenResponseTemplate, token,
                                        config_.value().get().token_ttl),
                        {{"Content-Type", "application/json"}});
    }
    return FilterHeadersStatus::Continue;
  }
  return rootCtx->checkAuthRule(
             [rootCtx, this](const auto& config, const auto& allow_set) {
               return rootCtx->checkPlugin(config, allow_set, route_name_);
             })
             ? FilterHeadersStatus::Continue
             : FilterHeadersStatus::StopIteration;
}

FilterDataStatus PluginContext::onRequestBody(size_t body_size,
                                              bool end_stream) {
  if (!check_body_params_) {
    return FilterDataStatus::Continue;
  }
  body_total_size_ += body_size;
  if (!end_stream) {
    return FilterDataStatus::StopIterationAndBuffer;
  }
  auto* rootCtx = rootContext();
  auto body =
      getBufferBytes(WasmBufferType::HttpRequestBody, 0, body_total_size_);
  LOG_DEBUG(absl::StrFormat("authorize request body: %s", body->toString()));
  std::string token, err_msg;
  if (rootCtx->generateToken(config_.value(), route_name_, body->view(), &token,
                             &err_msg)) {
    sendLocalResponse(200, "",
                      absl::StrFormat(TokenResponseTemplate, token,
                                      config_.value().get().token_ttl),
                      {{"Content-Type", "application/json"}});
    return FilterDataStatus::Continue;
  }
  sendLocalResponse(400, generateRcDetails(err_msg), err_msg, {});
  return FilterDataStatus::StopIterationNoBuffer;
}

#ifdef NULL_PLUGIN

}  // namespace oauth
}  // namespace null_plugin
}  // namespace proxy_wasm

#endif
