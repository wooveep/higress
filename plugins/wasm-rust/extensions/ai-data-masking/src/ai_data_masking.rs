// Copyright (c) 2025 Alibaba Group Holding Ltd.
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

use crate::deny_word::{DenyWord, MatchResult, MatchRule, MatchType as DenyMatchType};
use crate::msg_win_openai::MsgWindow;
use fancy_regex::Regex;
use grok::patterns;
use higress_wasm_rust::cluster_wrapper::{Cluster, K8sCluster};
use higress_wasm_rust::log::Log;
use higress_wasm_rust::plugin_wrapper::{HttpContextWrapper, RootContextWrapper};
use higress_wasm_rust::request_wrapper::{get_request_path, has_request_body};
use higress_wasm_rust::rule_matcher::{on_configure, RuleMatcher, SharedRuleMatcher};
use http::Method;
use jsonpath_rust::{JsonPath, JsonPathValue};
use lazy_static::lazy_static;
use multimap::MultiMap;
use proxy_wasm::hostcalls;
use proxy_wasm::traits::{Context, HttpContext, RootContext};
use proxy_wasm::types::{Bytes, ContextType, DataAction, HeaderAction, LogLevel};
use serde::de::Error;
use serde::{Deserialize, Deserializer};
use serde_json::{json, Value};
use std::cell::RefCell;
use std::collections::{BTreeMap, HashMap, VecDeque};
use std::fmt::Write;
use std::ops::DerefMut;
use std::rc::{Rc, Weak};
use std::str::FromStr;
use std::time::Duration;
use std::vec;

proxy_wasm::main! {{
    proxy_wasm::set_log_level(LogLevel::Trace);
    proxy_wasm::set_root_context(|_|Box::new(AiDataMaskingRoot::new()));
}}

const PLUGIN_NAME: &str = "ai-data-masking";
const GROK_PATTERN: &str = r"%\{(?<name>(?<pattern>[A-z0-9]+)(?::(?<alias>[A-z0-9_:;\/\s\.]+))?)\}";
const REQUEST_PHASE: &str = "request";
const RESPONSE_PHASE: &str = "response";

struct System {
    deny_word: DenyWord,
    grok_regex: Regex,
    grok_patterns: BTreeMap<String, String>,
}

lazy_static! {
    static ref SYSTEM: System = System::new();
}

struct AiDataMaskingRoot {
    log: Log,
    rule_matcher: SharedRuleMatcher<AiDataMaskingConfig>,
}

struct AiDataMasking {
    weak: Weak<RefCell<Box<dyn HttpContextWrapper<AiDataMaskingConfig>>>>,
    config: Option<Rc<RuntimeAiDataMaskingConfig>>,
    mask_map: HashMap<String, Option<String>>,
    is_openai: bool,
    is_openai_stream: Option<bool>,
    stream: bool,
    request_block_audit_reported: bool,
    log: Log,
    msg_window: MsgWindow,
    char_window_size: usize,
    byte_window_size: usize,
}

fn deserialize_regexp<'de, D>(deserializer: D) -> Result<Regex, D::Error>
where
    D: Deserializer<'de>,
{
    let value: Value = Deserialize::deserialize(deserializer)?;
    if let Some(pattern) = value.as_str() {
        let (expanded, _) = SYSTEM.grok_to_pattern(pattern);
        if let Ok(regex) = Regex::new(&expanded) {
            Ok(regex)
        } else if let Ok(regex) = Regex::new(pattern) {
            Ok(regex)
        } else {
            Err(Error::custom(format!("regexp error field {}", pattern)))
        }
    } else {
        Err(Error::custom("regexp error not string".to_string()))
    }
}

fn deserialize_type<'de, D>(deserializer: D) -> Result<ReplaceType, D::Error>
where
    D: Deserializer<'de>,
{
    let value: Value = Deserialize::deserialize(deserializer)?;
    if let Some(t) = value.as_str() {
        if t == "replace" {
            Ok(ReplaceType::Replace)
        } else if t == "hash" {
            Ok(ReplaceType::Hash)
        } else {
            Err(Error::custom(format!("regexp error value {}", t)))
        }
    } else {
        Err(Error::custom("type error not string".to_string()))
    }
}

fn deserialize_optional_string<'de, D>(deserializer: D) -> Result<Option<String>, D::Error>
where
    D: Deserializer<'de>,
{
    let value: Option<Value> = Option::deserialize(deserializer)?;
    Ok(match value {
        None | Some(Value::Null) => None,
        Some(Value::String(text)) => Some(text),
        Some(other) => Some(other.to_string()),
    })
}

fn deserialize_jsonpath<'de, D>(deserializer: D) -> Result<Vec<JsonPath>, D::Error>
where
    D: Deserializer<'de>,
{
    let value: Vec<String> = Deserialize::deserialize(deserializer)?;
    let mut ret = Vec::new();
    for item in value {
        if item.is_empty() {
            continue;
        }
        match JsonPath::from_str(&item) {
            Ok(jsonpath) => ret.push(jsonpath),
            Err(_) => return Err(Error::custom(format!("jsonpath error value {}", item))),
        }
    }
    Ok(ret)
}

#[derive(Debug, Clone, PartialEq, Eq)]
enum ReplaceType {
    Replace,
    Hash,
}

#[derive(Debug, Deserialize, Clone)]
struct ReplaceRule {
    #[allow(dead_code)]
    #[serde(default, deserialize_with = "deserialize_optional_string")]
    id: Option<String>,
    #[serde(
        alias = "pattern",
        alias = "regex",
        deserialize_with = "deserialize_regexp"
    )]
    regex: Regex,
    #[serde(
        alias = "replace_type",
        alias = "type",
        deserialize_with = "deserialize_type"
    )]
    type_: ReplaceType,
    #[serde(default)]
    restore: bool,
    #[serde(default, alias = "replace_value")]
    value: String,
    #[allow(dead_code)]
    #[serde(default)]
    description: String,
    #[serde(default)]
    priority: i32,
    #[serde(default = "default_true")]
    enabled: bool,
}

#[derive(Debug, Deserialize, Clone)]
struct DenyRule {
    #[serde(default, deserialize_with = "deserialize_optional_string")]
    id: Option<String>,
    pattern: String,
    #[serde(default = "default_match_type")]
    match_type: String,
    #[allow(dead_code)]
    #[serde(default)]
    description: String,
    #[serde(default)]
    priority: i32,
    #[serde(default = "default_true")]
    enabled: bool,
}

impl DenyRule {
    fn to_match_rule(&self) -> Result<MatchRule, String> {
        let pattern = self.pattern.trim();
        if pattern.is_empty() {
            return Err("pattern is empty".to_string());
        }
        let rule_id = self.id.clone();
        let priority = self.priority;
        let enabled = self.enabled;
        match self.match_type.to_lowercase().as_str() {
            "contains" => {
                let regex = Regex::new(&format!("(?i:{})", regex_escape(pattern)))
                    .map_err(|err| err.to_string())?;
                Ok(MatchRule::new(
                    rule_id,
                    pattern.to_string(),
                    DenyMatchType::Contains,
                    priority,
                    enabled,
                    Some(regex),
                ))
            }
            "exact" => Ok(MatchRule::new(
                rule_id,
                pattern.to_string(),
                DenyMatchType::Exact,
                priority,
                enabled,
                None,
            )),
            "regex" => {
                let (expanded, _) = SYSTEM.grok_to_pattern(pattern);
                let regex =
                    Regex::new(&format!("(?i:{})", expanded)).map_err(|err| err.to_string())?;
                Ok(MatchRule::new(
                    rule_id,
                    pattern.to_string(),
                    DenyMatchType::Regex,
                    priority,
                    enabled,
                    Some(regex),
                ))
            }
            other => Err(format!("unsupported match_type {}", other)),
        }
    }
}

#[derive(Default, Debug, Deserialize, Clone)]
struct AuditSinkConfig {
    #[serde(default)]
    service_name: String,
    #[serde(default)]
    namespace: String,
    #[serde(default)]
    port: u16,
    #[serde(default)]
    path: String,
    #[serde(default)]
    timeout_ms: u64,
}

impl AuditSinkConfig {
    fn is_enabled(&self) -> bool {
        !self.service_name.trim().is_empty() && self.port > 0
    }

    fn cluster(&self) -> K8sCluster {
        let namespace = if self.namespace.trim().is_empty() {
            "default"
        } else {
            self.namespace.trim()
        };
        K8sCluster::new(
            self.service_name.trim(),
            namespace,
            &self.port.to_string(),
            "",
            "",
        )
    }

    fn url(&self) -> String {
        let path = if self.path.starts_with('/') {
            self.path.clone()
        } else {
            format!("/{}", self.path)
        };
        format!("http://{}{}", self.cluster().host_name(), path)
    }

    fn timeout(&self) -> Duration {
        Duration::from_millis(if self.timeout_ms == 0 {
            2000
        } else {
            self.timeout_ms
        })
    }
}

fn default_deny_openai() -> bool {
    true
}

fn default_deny_raw() -> bool {
    false
}

fn default_system_deny() -> bool {
    false
}

fn default_deny_code() -> u16 {
    200
}

fn default_deny_content_type() -> String {
    "application/json".to_string()
}

fn default_deny_raw_message() -> String {
    "{\"errmsg\":\"提问或回答中包含敏感词，已被屏蔽\"}".to_string()
}

fn default_deny_message() -> String {
    "提问或回答中包含敏感词，已被屏蔽".to_string()
}

fn default_match_type() -> String {
    "contains".to_string()
}

fn default_true() -> bool {
    true
}

#[derive(Default, Debug, Deserialize, Clone)]
pub struct AiDataMaskingConfig {
    #[serde(default = "default_deny_openai")]
    deny_openai: bool,
    #[serde(default = "default_deny_raw")]
    deny_raw: bool,
    #[serde(default, deserialize_with = "deserialize_jsonpath")]
    deny_jsonpath: Vec<JsonPath>,
    #[serde(default = "default_system_deny")]
    system_deny: bool,
    #[serde(default)]
    system_deny_words: Option<Vec<String>>,
    #[serde(default = "default_deny_code")]
    deny_code: u16,
    #[serde(default = "default_deny_message")]
    deny_message: String,
    #[serde(default = "default_deny_raw_message")]
    deny_raw_message: String,
    #[serde(default = "default_deny_content_type")]
    deny_content_type: String,
    #[serde(default)]
    replace_roles: Vec<ReplaceRule>,
    #[serde(default)]
    replace_rules: Vec<ReplaceRule>,
    #[serde(default)]
    deny_words: Vec<String>,
    #[serde(default)]
    deny_rules: Vec<DenyRule>,
    #[serde(default)]
    audit_sink: Option<AuditSinkConfig>,
}

#[derive(Debug, Clone)]
struct RuntimeAiDataMaskingConfig {
    deny_openai: bool,
    deny_raw: bool,
    deny_jsonpath: Vec<JsonPath>,
    deny_code: u16,
    deny_message: String,
    deny_raw_message: String,
    deny_content_type: String,
    replace_rules: Vec<ReplaceRule>,
    deny_matcher: DenyWord,
    audit_sink: Option<AuditSinkConfig>,
}

impl RuntimeAiDataMaskingConfig {
    fn from_config(raw: &AiDataMaskingConfig, log: &Log) -> Self {
        let mut deny_matcher = DenyWord::empty();
        let mut compiled_rules = Vec::new();
        for rule in &raw.deny_rules {
            match rule.to_match_rule() {
                Ok(compiled) => compiled_rules.push(compiled),
                Err(err) => log.warn(&format!(
                    "skip invalid deny rule pattern={}, reason={}",
                    rule.pattern, err
                )),
            }
        }
        for word in &raw.deny_words {
            let pattern = word.trim();
            if pattern.is_empty() {
                continue;
            }
            match Regex::new(&format!("(?i:{})", regex_escape(pattern))) {
                Ok(regex) => compiled_rules.push(MatchRule::new(
                    None,
                    pattern.to_string(),
                    DenyMatchType::Contains,
                    0,
                    true,
                    Some(regex),
                )),
                Err(err) => log.warn(&format!(
                    "skip invalid legacy deny word pattern={}, reason={}",
                    pattern, err
                )),
            }
        }
        deny_matcher.append(DenyWord::from_rules(compiled_rules));
        if raw.system_deny {
            match raw.system_deny_words.as_ref() {
                Some(system_deny_words) => {
                    let system_text = system_deny_words
                        .iter()
                        .map(|word| word.trim())
                        .filter(|word| !word.is_empty())
                        .collect::<Vec<_>>()
                        .join("\n");
                    deny_matcher.append(DenyWord::from_system_text(&system_text));
                }
                None => deny_matcher.append(SYSTEM.deny_word.clone()),
            }
        }

        let mut replace_rules = raw.replace_rules.clone();
        replace_rules.extend(raw.replace_roles.clone());
        replace_rules.retain(|rule| rule.enabled);
        replace_rules.sort_by(|left, right| right.priority.cmp(&left.priority));

        RuntimeAiDataMaskingConfig {
            deny_openai: raw.deny_openai,
            deny_raw: raw.deny_raw,
            deny_jsonpath: raw.deny_jsonpath.clone(),
            deny_code: raw.deny_code,
            deny_message: raw.deny_message.clone(),
            deny_raw_message: raw.deny_raw_message.clone(),
            deny_content_type: raw.deny_content_type.clone(),
            replace_rules,
            deny_matcher,
            audit_sink: raw.audit_sink.clone(),
        }
    }

    fn check_message(&self, message: &str, log: &Log) -> Option<MatchResult> {
        if let Some(result) = self.deny_matcher.check(message) {
            log.warn(&format!(
                "deny rule pattern={} match_type={} matched_text={}",
                result.pattern,
                result.match_type.as_str(),
                result.matched_text
            ));
            return Some(result);
        }
        None
    }
}

static SYSTEM_PATTERNS: &[(&str, &str)] = &[
    ("MOBILE", r#"\d{8,11}"#),
    ("IDCARD", r#"\d{17}[0-9xX]|\d{15}"#),
];

impl System {
    fn new() -> Self {
        let grok_regex = Regex::new(GROK_PATTERN).unwrap();
        let grok_patterns = BTreeMap::new();
        let mut system = System {
            deny_word: DenyWord::system(),
            grok_regex,
            grok_patterns,
        };
        system.init();
        system
    }

    fn init(&mut self) {
        let mut grok_temp_patterns = VecDeque::new();
        for pattern_group in [patterns(), SYSTEM_PATTERNS] {
            for &(key, value) in pattern_group {
                if self.grok_regex.is_match(value).is_ok_and(|matched| matched) {
                    grok_temp_patterns.push_back((String::from(key), String::from(value)));
                } else {
                    self.grok_patterns
                        .insert(String::from(key), String::from(value));
                }
            }
        }
        let mut last_ok: Option<String> = None;
        while let Some((key, value)) = grok_temp_patterns.pop_front() {
            if let Some(last_key) = &last_ok {
                if last_key == &key {
                    break;
                }
            }
            let (expanded, ok) = self.grok_to_pattern(&value);
            if ok {
                self.grok_patterns.insert(key, expanded);
                last_ok = None;
            } else {
                if last_ok.is_none() {
                    last_ok = Some(key.clone());
                }
                grok_temp_patterns.push_back((key, expanded));
            }
        }
    }

    fn grok_to_pattern(&self, pattern: &str) -> (String, bool) {
        let mut ok = true;
        let mut ret = pattern.to_string();
        for capture in self.grok_regex.captures_iter(pattern) {
            if capture.is_err() {
                ok = false;
                continue;
            }
            let capture = capture.unwrap();
            if let (Some(full), Some(name)) = (capture.get(0), capture.name("pattern")) {
                if let Some(found) = self.grok_patterns.get(name.as_str()) {
                    if let Some(alias) = capture.name("alias") {
                        ret = ret
                            .replace(full.as_str(), &format!("(?P<{}>{})", alias.as_str(), found));
                    } else {
                        ret = ret.replace(full.as_str(), found);
                    }
                } else {
                    ok = false;
                }
            }
        }
        (ret, ok)
    }
}

impl AiDataMaskingRoot {
    fn new() -> Self {
        AiDataMaskingRoot {
            log: Log::new(PLUGIN_NAME.to_string()),
            rule_matcher: Rc::new(RefCell::new(RuleMatcher::default())),
        }
    }
}

impl Context for AiDataMaskingRoot {}

impl RootContext for AiDataMaskingRoot {
    fn on_configure(&mut self, plugin_configuration_size: usize) -> bool {
        on_configure(
            self,
            plugin_configuration_size,
            self.rule_matcher.borrow_mut().deref_mut(),
            &self.log,
        )
    }

    fn create_http_context(&self, context_id: u32) -> Option<Box<dyn HttpContext>> {
        self.create_http_context_use_wrapper(context_id)
    }

    fn get_type(&self) -> Option<ContextType> {
        Some(ContextType::HttpContext)
    }
}

impl RootContextWrapper<AiDataMaskingConfig> for AiDataMaskingRoot {
    fn rule_matcher(&self) -> &SharedRuleMatcher<AiDataMaskingConfig> {
        &self.rule_matcher
    }

    fn create_http_context_wrapper(
        &self,
        _context_id: u32,
    ) -> Option<Box<dyn HttpContextWrapper<AiDataMaskingConfig>>> {
        Some(Box::new(AiDataMasking {
            weak: Weak::default(),
            mask_map: HashMap::new(),
            config: None,
            is_openai: false,
            is_openai_stream: None,
            stream: false,
            request_block_audit_reported: false,
            msg_window: MsgWindow::default(),
            log: Log::new(PLUGIN_NAME.to_string()),
            char_window_size: 0,
            byte_window_size: 0,
        }))
    }
}

impl AiDataMasking {
    fn check_message(&self, message: &str) -> Option<MatchResult> {
        self.config
            .as_ref()
            .and_then(|config| config.check_message(message, self.log()))
    }

    fn msg_to_response(&self, msg: &str, raw_msg: &str, content_type: &str) -> (String, String) {
        if !self.is_openai {
            (raw_msg.to_string(), content_type.to_string())
        } else if self.stream {
            (
                format!(
                    "data:{}\n\n",
                    json!({"choices": [{"index": 0, "delta": {"role": "assistant", "content": msg}}], "usage": {}})
                ),
                "text/event-stream;charset=UTF-8".to_string(),
            )
        } else {
            (
                json!({"choices": [{"index": 0, "message": {"role": "assistant", "content": msg}}], "usage": {}})
                    .to_string(),
                "application/json".to_string(),
            )
        }
    }

    fn deny(
        &mut self,
        in_response: bool,
        match_result: Option<&MatchResult>,
        request_phase: &str,
    ) -> DataAction {
        if self.should_report_block_audit(request_phase) {
            self.report_block_audit(match_result, request_phase);
        }
        let (deny_code, (deny_message, content_type)) = if let Some(config) = &self.config {
            (
                config.deny_code,
                self.msg_to_response(
                    &config.deny_message,
                    &config.deny_raw_message,
                    &config.deny_content_type,
                ),
            )
        } else {
            (
                default_deny_code(),
                self.msg_to_response(
                    &default_deny_message(),
                    &default_deny_raw_message(),
                    &default_deny_content_type(),
                ),
            )
        };
        if in_response {
            self.replace_http_response_body(deny_message.as_bytes());
            return DataAction::Continue;
        }
        self.send_http_response(
            deny_code as u32,
            vec![("Content-Type", &content_type)],
            Some(deny_message.as_bytes()),
        );
        DataAction::StopIterationAndBuffer
    }

    fn replace_request_msg(&mut self, message: &str) -> String {
        let replace_rules = match &self.config {
            Some(config) => config.replace_rules.clone(),
            None => return message.to_string(),
        };
        let mut msg = message.to_string();
        for rule in &replace_rules {
            msg = self.apply_replace_rule(&msg, rule);
        }
        if msg != message {
            self.log()
                .debug(&format!("replace_request_msg from {} to {}", message, msg));
        }
        msg
    }

    fn apply_replace_rule(&mut self, message: &str, rule: &ReplaceRule) -> String {
        let mut result = String::with_capacity(message.len());
        let mut last_end = 0usize;
        let mut changed = false;
        for matched in rule.regex.find_iter(message) {
            let Ok(matched) = matched else {
                continue;
            };
            let start = matched.start();
            let end = matched.end();
            if start < last_end {
                continue;
            }
            let from_word = matched.as_str();
            let to_word = self.build_replace_word(from_word, rule);
            result.push_str(&message[last_end..start]);
            result.push_str(&to_word);
            self.record_replacement(from_word, &to_word, rule);
            last_end = end;
            changed = true;
        }
        if !changed {
            return message.to_string();
        }
        result.push_str(&message[last_end..]);
        result
    }

    fn build_replace_word(&self, from_word: &str, rule: &ReplaceRule) -> String {
        match rule.type_ {
            ReplaceType::Hash => {
                let digest = hmac_sha256::Hash::hash(from_word.as_bytes());
                digest.iter().fold(String::new(), |mut output, byte| {
                    let _ = write!(output, "{byte:02x}");
                    output
                })
            }
            ReplaceType::Replace => {
                mask_numeric_middle(from_word, &rule.value)
                    .unwrap_or_else(|| rule.regex.replace(from_word, &rule.value).to_string())
            }
        }
    }

    fn record_replacement(&mut self, from_word: &str, to_word: &str, rule: &ReplaceRule) {
        if to_word.len() > self.byte_window_size {
            self.byte_window_size = to_word.len();
        }
        if to_word.chars().count() > self.char_window_size {
            self.char_window_size = to_word.chars().count();
        }
        if rule.restore && !to_word.is_empty() {
            match self.mask_map.entry(to_word.to_string()) {
                std::collections::hash_map::Entry::Occupied(mut existed) => {
                    existed.insert(None);
                }
                std::collections::hash_map::Entry::Vacant(entry) => {
                    entry.insert(Some(from_word.to_string()));
                }
            }
        }
    }

    fn restore_response_msg(&self, message: &str) -> String {
        if self.mask_map.is_empty() {
            return message.to_string();
        }
        let mut msg = message.to_string();
        for (from_word, to_word) in self.mask_map.iter() {
            if let Some(origin) = to_word {
                msg = msg.replace(from_word, origin);
            }
        }
        msg
    }

    fn process_request_string(&mut self, content: &mut String) -> Option<MatchResult> {
        if let Some(match_result) = self.check_message(content) {
            return Some(match_result);
        }
        let new_content = self.replace_request_msg(content);
        if new_content != *content {
            *content = new_content;
        }
        None
    }

    fn process_response_string(&mut self, content: &mut String) -> Option<MatchResult> {
        let new_content = self.restore_response_msg(content);
        if new_content != *content {
            *content = new_content;
        }
        None
    }

    fn process_request_text_value(&mut self, value: &mut Value) -> Option<MatchResult> {
        match value {
            Value::String(content) => self.process_request_string(content),
            Value::Array(items) => {
                for item in items {
                    if let Some(match_result) = self.process_request_text_value(item) {
                        return Some(match_result);
                    }
                }
                None
            }
            Value::Object(object) => {
                for key in ["text", "input_text", "content", "reasoning_content"] {
                    if let Some(item) = object.get_mut(key) {
                        if let Some(match_result) = self.process_request_text_value(item) {
                            return Some(match_result);
                        }
                    }
                }
                None
            }
            _ => None,
        }
    }

    fn process_response_text_value(&mut self, value: &mut Value) -> Option<MatchResult> {
        match value {
            Value::String(content) => self.process_response_string(content),
            Value::Array(items) => {
                for item in items {
                    if let Some(match_result) = self.process_response_text_value(item) {
                        return Some(match_result);
                    }
                }
                None
            }
            Value::Object(object) => {
                for key in ["text", "output_text", "content", "reasoning_content"] {
                    if let Some(item) = object.get_mut(key) {
                        if let Some(match_result) = self.process_response_text_value(item) {
                            return Some(match_result);
                        }
                    }
                }
                None
            }
            _ => None,
        }
    }

    fn is_user_role(value: &Value) -> bool {
        value
            .get("role")
            .and_then(Value::as_str)
            .map(|role| role.eq_ignore_ascii_case("user"))
            .unwrap_or(true)
    }

    fn is_input_turn_start(value: &Value) -> bool {
        if !value.is_object() {
            return true;
        }
        Self::is_user_role(value)
    }

    fn should_report_block_audit(&mut self, request_phase: &str) -> bool {
        match request_phase {
            REQUEST_PHASE => {
                if self.request_block_audit_reported {
                    false
                } else {
                    self.request_block_audit_reported = true;
                    true
                }
            }
            RESPONSE_PHASE => false,
            _ => true,
        }
    }

    fn process_openai_request_body(&mut self, value: &mut Value) -> Option<MatchResult> {
        let object = match value.as_object_mut() {
            Some(object) => object,
            None => return None,
        };
        let has_openai_shape = object.contains_key("messages")
            || object.contains_key("input")
            || object.contains_key("system");
        if !has_openai_shape {
            return None;
        }
        self.is_openai = true;
        self.stream = object
            .get("stream")
            .and_then(Value::as_bool)
            .unwrap_or(false);

        if let Some(system) = object.get_mut("system") {
            if let Some(match_result) = self.process_request_text_value(system) {
                return Some(match_result);
            }
        }

        if let Some(messages) = object.get_mut("messages").and_then(Value::as_array_mut) {
            if let Some(match_result) = process_turns_with_history_cleanup(
                messages,
                Self::is_user_role,
                |message| self.process_request_text_value(message),
                |_| {},
            ) {
                return Some(match_result);
            }
        }

        if let Some(input) = object.get_mut("input") {
            match input {
                Value::Array(items) => {
                    if let Some(match_result) = process_turns_with_history_cleanup(
                        items,
                        Self::is_input_turn_start,
                        |item| self.process_request_text_value(item),
                        |_| {},
                    ) {
                        return Some(match_result);
                    }
                }
                other => {
                    if let Some(match_result) = self.process_request_text_value(other) {
                        return Some(match_result);
                    }
                }
            }
        }

        None
    }

    fn process_openai_response_body(&mut self, value: &mut Value) {
        if let Some(choices) = value.get_mut("choices").and_then(Value::as_array_mut) {
            for choice in choices {
                if let Some(message) = choice.get_mut("message") {
                    let _ = self.process_response_text_value(message);
                }
                if let Some(delta) = choice.get_mut("delta") {
                    let _ = self.process_response_text_value(delta);
                }
            }
        }
    }

    fn process_request_jsonpath(
        &mut self,
        json_value: &Value,
        raw_body: &mut String,
    ) -> Option<MatchResult> {
        let jsonpaths = match &self.config {
            Some(config) => config,
            None => return None,
        }
        .deny_jsonpath
        .clone();
        for jsonpath in jsonpaths {
            for value in jsonpath.find_slice(json_value) {
                if let JsonPathValue::Slice(data, _) = value {
                    if let Some(text) = data.as_str() {
                        if let Some(match_result) = self.check_message(text) {
                            return Some(match_result);
                        }
                        let content = text.to_string();
                        let new_content = self.replace_request_msg(&content);
                        if new_content != content {
                            *raw_body = raw_body.replace(
                                &Value::String(content).to_string(),
                                &Value::String(new_content).to_string(),
                            );
                        }
                    }
                }
            }
        }
        None
    }

    fn report_block_audit(&mut self, match_result: Option<&MatchResult>, request_phase: &str) {
        let config = match &self.config {
            Some(config) => config,
            None => return,
        };
        let audit_sink = match &config.audit_sink {
            Some(audit_sink) if audit_sink.is_enabled() => audit_sink.clone(),
            _ => return,
        };

        let request_id = self
            .get_http_request_header("x-request-id")
            .or_else(|| self.get_http_request_header("x-envoy-request-id"))
            .unwrap_or_default();
        let consumer_name = self
            .get_http_request_header("x-mse-consumer")
            .or_else(|| self.get_http_request_header("X-Mse-Consumer"))
            .unwrap_or_default();
        let route_name = get_string_property(&["route_name"]).unwrap_or_default();

        let blocked_reason = json!({
            "blocked_by": "sensitive_word",
            "request_phase": request_phase,
            "route_name": route_name,
            "consumer_name": consumer_name,
            "match_type": match_result.map(|result| result.match_type.as_str()).unwrap_or_default(),
            "matched_rule": match_result.map(|result| result.pattern.clone()).unwrap_or_default(),
            "matched_excerpt": match_result.map(|result| result.excerpt.clone()).unwrap_or_default(),
        })
        .to_string();

        let payload = json!({
            "requestId": empty_to_null(&request_id),
            "routeName": empty_to_null(&route_name),
            "consumerName": empty_to_null(&consumer_name),
            "blockedBy": "sensitive_word",
            "requestPhase": request_phase,
            "blockedReasonJson": blocked_reason,
            "matchType": match_result.map(|result| result.match_type.as_str()),
            "matchedRule": match_result.map(|result| result.pattern.clone()),
            "matchedExcerpt": match_result.map(|result| result.excerpt.clone()),
        })
        .to_string();

        let mut headers = MultiMap::new();
        headers.insert("Content-Type".to_string(), "application/json".to_string());
        let cluster = audit_sink.cluster();
        let audit_url = audit_sink.url();
        let cluster_name = cluster.cluster_name();
        let callback_phase = request_phase.to_string();
        let callback_request_id = request_id.clone();
        let callback_cluster_name = cluster_name.clone();
        let callback_audit_url = audit_url.clone();
        if let Err(status) = self.http_call(
            &cluster,
            &Method::POST,
            &audit_url,
            headers,
            Some(payload.as_bytes()),
            Box::new(move |status_code, response_headers, response_body| {
                if (200..300).contains(&status_code) {
                    return;
                }
                let response_body = response_body
                    .map(|body| String::from_utf8_lossy(&body).to_string())
                    .unwrap_or_default();
                let response_status = response_headers
                    .get(":status")
                    .cloned()
                    .unwrap_or_default();
                let log = Log::new(PLUGIN_NAME.to_string());
                log.warn(&format!(
                    "failed to report sensitive block audit, phase={}, request_id={}, cluster={}, url={}, http_status={}, header_status={}, body={}",
                    callback_phase,
                    callback_request_id,
                    callback_cluster_name,
                    callback_audit_url,
                    status_code,
                    response_status,
                    response_body
                ));
            }),
            audit_sink.timeout(),
        ) {
            self.log.warn(&format!(
                "failed to dispatch sensitive block audit, phase={}, request_id={}, cluster={}, url={}, status={:?}",
                request_phase,
                request_id,
                cluster_name,
                audit_url,
                status
            ));
        }
    }
}

impl Context for AiDataMasking {}

impl HttpContext for AiDataMasking {
    fn on_http_request_headers(
        &mut self,
        _num_headers: usize,
        _end_of_stream: bool,
    ) -> HeaderAction {
        if has_request_body() {
            self.set_http_request_header("Content-Length", None);
            HeaderAction::StopIteration
        } else {
            HeaderAction::Continue
        }
    }

    fn on_http_response_headers(
        &mut self,
        _num_headers: usize,
        _end_of_stream: bool,
    ) -> HeaderAction {
        self.set_http_response_header("Content-Length", None);
        HeaderAction::Continue
    }

    fn on_http_response_body(&mut self, body_size: usize, end_of_stream: bool) -> DataAction {
        if !self.stream {
            return DataAction::Continue;
        }
        if body_size > 0 {
            if let Some(body) = self.get_http_response_body(0, body_size) {
                if self.is_openai && self.is_openai_stream.is_none() {
                    self.is_openai_stream = Some(body.starts_with(b"data:"));
                }
                self.msg_window
                    .push(&body, self.is_openai_stream.unwrap_or_default());
                let mask_map = self.mask_map.clone();
                for message in self.msg_window.messages_iter_mut() {
                    if let Ok(mut msg) = String::from_utf8(message.clone()) {
                        msg = restore_with_mask_map(&msg, &mask_map);
                        message.clear();
                        message.extend_from_slice(msg.as_bytes());
                    }
                }
            }
        }
        let new_body = if end_of_stream {
            self.msg_window
                .finish(self.is_openai_stream.unwrap_or_default())
        } else {
            self.msg_window.pop(
                self.char_window_size * 2,
                self.byte_window_size * 2,
                self.is_openai_stream.unwrap_or_default(),
            )
        };
        self.replace_http_response_body(&new_body);
        DataAction::Continue
    }
}

impl HttpContextWrapper<AiDataMaskingConfig> for AiDataMasking {
    fn init_self_weak(
        &mut self,
        self_weak: Weak<RefCell<Box<dyn HttpContextWrapper<AiDataMaskingConfig>>>>,
    ) {
        self.weak = self_weak;
    }

    fn log(&self) -> &Log {
        &self.log
    }

    fn on_config(&mut self, config: Rc<AiDataMaskingConfig>) {
        self.config = Some(Rc::new(RuntimeAiDataMaskingConfig::from_config(
            config.as_ref(),
            &self.log,
        )));
    }

    fn cache_request_body(&self) -> bool {
        true
    }

    fn cache_response_body(&self) -> bool {
        !self.stream
    }

    fn on_http_request_complete_body(&mut self, req_body: &Bytes) -> DataAction {
        let config = match &self.config {
            Some(config) => config.clone(),
            None => return DataAction::Continue,
        };
        if get_request_path().contains("count_tokens") {
            return DataAction::Continue;
        }

        let mut body_string = match serde_json::from_slice::<Value>(req_body) {
            Ok(value) => value.to_string(),
            Err(_) => match String::from_utf8(req_body.clone()) {
                Ok(body) => body,
                Err(_) => return DataAction::Continue,
            },
        };

        let mut handled_json = false;
        if let Ok(mut json_value) = serde_json::from_str::<Value>(&body_string) {
            if config.deny_openai {
                if let Some(match_result) = self.process_openai_request_body(&mut json_value) {
                    return self.deny(false, Some(&match_result), REQUEST_PHASE);
                }
                handled_json = self.is_openai;
            }
            body_string = json_value.to_string();

            if !config.deny_jsonpath.is_empty() {
                if let Some(match_result) =
                    self.process_request_jsonpath(&json_value, &mut body_string)
                {
                    return self.deny(false, Some(&match_result), REQUEST_PHASE);
                }
                handled_json = true;
            }
        }

        if config.deny_raw {
            if let Some(match_result) = self.check_message(&body_string) {
                return self.deny(false, Some(&match_result), REQUEST_PHASE);
            }
            body_string = self.replace_request_msg(&body_string);
        }

        if handled_json || config.deny_raw {
            self.replace_http_request_body(body_string.as_bytes());
        }
        DataAction::Continue
    }

    fn on_http_response_complete_body(&mut self, res_body: &Bytes) -> DataAction {
        let config = match &self.config {
            Some(config) => config.clone(),
            None => return DataAction::Continue,
        };

        let mut body_string = match serde_json::from_slice::<Value>(res_body) {
            Ok(value) => value.to_string(),
            Err(_) => match String::from_utf8(res_body.clone()) {
                Ok(body) => body,
                Err(_) => return DataAction::Continue,
            },
        };

        let mut handled_json = false;
        if self.is_openai {
            if let Ok(mut json_value) = serde_json::from_str::<Value>(&body_string) {
                self.process_openai_response_body(&mut json_value);
                body_string = json_value.to_string();
                handled_json = true;
            }
        }

        if config.deny_raw {
            body_string = self.restore_response_msg(&body_string);
            handled_json = true;
        }

        if handled_json {
            self.replace_http_response_body(body_string.as_bytes());
        }
        DataAction::Continue
    }
}

fn regex_escape(value: &str) -> String {
    let mut escaped = String::with_capacity(value.len());
    for ch in value.chars() {
        match ch {
            '\\' | '.' | '+' | '*' | '?' | '(' | ')' | '|' | '[' | ']' | '{' | '}' | '^' | '$' => {
                escaped.push('\\');
                escaped.push(ch);
            }
            _ => escaped.push(ch),
        }
    }
    escaped
}

fn get_string_property(path: &[&str]) -> Option<String> {
    if let Ok(Some(bytes)) = hostcalls::get_property(path.to_vec()) {
        if let Ok(value) = String::from_utf8(bytes) {
            let trimmed = value.trim();
            if !trimmed.is_empty() {
                return Some(trimmed.to_string());
            }
        }
    }
    None
}

fn empty_to_null(value: &str) -> Value {
    if value.trim().is_empty() {
        Value::Null
    } else {
        Value::String(value.to_string())
    }
}

fn restore_with_mask_map(message: &str, mask_map: &HashMap<String, Option<String>>) -> String {
    if mask_map.is_empty() {
        return message.to_string();
    }
    let mut restored = message.to_string();
    for (masked, origin) in mask_map.iter() {
        if let Some(original) = origin {
            restored = restored.replace(masked, original);
        }
    }
    restored
}

fn find_next_turn_index<F>(items: &[Value], start: usize, is_turn_start: F) -> Option<usize>
where
    F: Fn(&Value) -> bool + Copy,
{
    items
        .iter()
        .enumerate()
        .skip(start)
        .find(|(_, item)| is_turn_start(item))
        .map(|(idx, _)| idx)
}

fn process_turns_with_history_cleanup<F, P, H>(
    items: &mut Vec<Value>,
    is_turn_start: F,
    mut process_turn: P,
    mut on_historical_match: H,
) -> Option<MatchResult>
where
    F: Fn(&Value) -> bool + Copy,
    P: FnMut(&mut Value) -> Option<MatchResult>,
    H: FnMut(&MatchResult),
{
    let latest_turn_idx = items.iter().rposition(|item| is_turn_start(item));
    let Some(latest_turn_idx) = latest_turn_idx else {
        return None;
    };

    if let Some(match_result) = process_turn(&mut items[latest_turn_idx]) {
        return Some(match_result);
    }

    if latest_turn_idx == 0 {
        return None;
    }

    let mut idx = latest_turn_idx;
    while idx > 0 {
        idx -= 1;
        if !is_turn_start(&items[idx]) {
            continue;
        }
        if let Some(match_result) = process_turn(&mut items[idx]) {
            on_historical_match(&match_result);
            let drain_end =
                find_next_turn_index(items, idx + 1, is_turn_start).unwrap_or(items.len());
            items.drain(idx..drain_end);
        }
    }

    None
}

fn mask_numeric_middle(from_word: &str, replace_value: &str) -> Option<String> {
    let mut chars = from_word.chars().collect::<Vec<_>>();
    if chars.len() < 7 || !chars.iter().all(|ch| ch.is_ascii_digit()) {
        return None;
    }
    let mut replace_chars = replace_value.chars();
    let fill = replace_chars.next()?;
    if replace_chars.next().is_some() {
        return None;
    }
    let prefix_len = 3usize.min(chars.len().saturating_sub(1));
    let suffix_len = 1usize;
    let middle_end = chars.len().saturating_sub(suffix_len);
    if prefix_len >= middle_end {
        return None;
    }
    for ch in chars.iter_mut().take(middle_end).skip(prefix_len) {
        *ch = fill;
    }
    Some(chars.into_iter().collect())
}

#[cfg(test)]
mod tests {
    use super::*;

    fn contains_nanjing(value: &mut Value) -> Option<MatchResult> {
        let contains = match value {
            Value::String(text) => text.contains("南京"),
            Value::Array(items) => items
                .iter_mut()
                .any(|item| contains_nanjing(item).is_some()),
            Value::Object(object) => object
                .values_mut()
                .any(|item| contains_nanjing(item).is_some()),
            _ => false,
        };

        contains.then(|| MatchResult {
            rule_id: None,
            pattern: "南京".to_string(),
            match_type: DenyMatchType::Contains,
            matched_text: "南京".to_string(),
            excerpt: "南京".to_string(),
        })
    }

    #[test]
    fn process_turns_with_history_cleanup_removes_historical_offending_message_group() {
        let mut messages = vec![
            json!({"role": "user", "content": "请介绍南京的旅游景点"}),
            json!({"role": "assistant", "content": "提问或回答中包含敏感词，已被屏蔽"}),
            json!({"role": "user", "content": "请介绍北京的旅游景点"}),
        ];

        let result = process_turns_with_history_cleanup(
            &mut messages,
            AiDataMasking::is_user_role,
            contains_nanjing,
            |_| {},
        );

        assert!(result.is_none());
        assert_eq!(messages.len(), 1);
        assert_eq!(messages[0]["content"], "请介绍北京的旅游景点");
    }

    #[test]
    fn process_turns_with_history_cleanup_keeps_latest_turn_blocking() {
        let mut messages = vec![
            json!({"role": "user", "content": "你好"}),
            json!({"role": "assistant", "content": "你好，请问有什么可以帮你"}),
            json!({"role": "user", "content": "请介绍南京的旅游景点"}),
        ];

        let result = process_turns_with_history_cleanup(
            &mut messages,
            AiDataMasking::is_user_role,
            contains_nanjing,
            |_| {},
        );

        assert!(result.is_some());
        assert_eq!(messages.len(), 3);
    }

    #[test]
    fn process_turns_with_history_cleanup_removes_historical_input_group() {
        let mut input = vec![
            json!({"role": "user", "content": "南京怎么样"}),
            json!({"role": "assistant", "content": "提问或回答中包含敏感词，已被屏蔽"}),
            json!({"role": "user", "content": "北京怎么样"}),
        ];

        let result = process_turns_with_history_cleanup(
            &mut input,
            AiDataMasking::is_input_turn_start,
            contains_nanjing,
            |_| {},
        );

        assert!(result.is_none());
        assert_eq!(input.len(), 1);
        assert_eq!(input[0]["content"], "北京怎么样");
    }

    #[test]
    fn should_report_block_audit_only_once_per_phase() {
        let mut masking = AiDataMasking {
            weak: Weak::default(),
            config: None,
            mask_map: HashMap::new(),
            is_openai: false,
            is_openai_stream: None,
            stream: false,
            request_block_audit_reported: false,
            log: Log::new(PLUGIN_NAME.to_string()),
            msg_window: MsgWindow::default(),
            char_window_size: 0,
            byte_window_size: 0,
        };

        assert!(masking.should_report_block_audit(REQUEST_PHASE));
        assert!(!masking.should_report_block_audit(REQUEST_PHASE));
        assert!(!masking.should_report_block_audit(RESPONSE_PHASE));
    }

    #[test]
    fn system_deny_defaults_to_disabled() {
        let config = RuntimeAiDataMaskingConfig::from_config(
            &AiDataMaskingConfig::default(),
            &Log::new(PLUGIN_NAME.to_string()),
        );

        assert!(config
            .check_message("天安门", &Log::new(PLUGIN_NAME.to_string()))
            .is_none());
    }

    #[test]
    fn custom_system_deny_words_override_bundled_dictionary() {
        let config = RuntimeAiDataMaskingConfig::from_config(
            &AiDataMaskingConfig {
                system_deny: true,
                system_deny_words: Some(vec!["自定义词".to_string()]),
                ..AiDataMaskingConfig::default()
            },
            &Log::new(PLUGIN_NAME.to_string()),
        );

        assert!(config
            .check_message("自定义词", &Log::new(PLUGIN_NAME.to_string()))
            .is_some());
        assert!(config
            .check_message("天安门", &Log::new(PLUGIN_NAME.to_string()))
            .is_none());
    }

    #[test]
    fn bundled_system_dictionary_is_used_when_projection_words_are_absent() {
        let config = RuntimeAiDataMaskingConfig::from_config(
            &AiDataMaskingConfig {
                system_deny: true,
                system_deny_words: None,
                ..AiDataMaskingConfig::default()
            },
            &Log::new(PLUGIN_NAME.to_string()),
        );

        assert!(config
            .check_message("天安门", &Log::new(PLUGIN_NAME.to_string()))
            .is_some());
    }

    #[test]
    fn short_system_dictionary_terms_do_not_block_benign_longer_queries() {
        let config = RuntimeAiDataMaskingConfig::from_config(
            &AiDataMaskingConfig {
                system_deny: true,
                system_deny_words: Some(vec!["天安门".to_string(), "天安门事件".to_string()]),
                ..AiDataMaskingConfig::default()
            },
            &Log::new(PLUGIN_NAME.to_string()),
        );

        assert!(config
            .check_message("天安门的景点", &Log::new(PLUGIN_NAME.to_string()))
            .is_none());
        assert!(config
            .check_message("请介绍天安门事件", &Log::new(PLUGIN_NAME.to_string()))
            .is_some());
    }

    #[test]
    fn mask_numeric_middle_keeps_phone_prefix_and_suffix() {
        assert_eq!(
            Some("19011111112".to_string()),
            mask_numeric_middle("19028732382", "1")
        );
        assert_eq!(None, mask_numeric_middle("foo@example.com", "1"));
        assert_eq!(None, mask_numeric_middle("19028732382", "11"));
    }
}
