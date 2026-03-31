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

use fancy_regex::Regex;
use rust_embed::Embed;

#[derive(Embed)]
#[folder = "res/"]
struct Asset;

#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub(crate) enum MatchType {
    Contains,
    Exact,
    Regex,
}

impl MatchType {
    pub(crate) fn as_str(&self) -> &'static str {
        match self {
            MatchType::Contains => "contains",
            MatchType::Exact => "exact",
            MatchType::Regex => "regex",
        }
    }
}

#[derive(Debug, Clone)]
pub(crate) struct MatchRule {
    pub(crate) rule_id: Option<String>,
    pub(crate) pattern: String,
    pub(crate) normalized_pattern: String,
    pub(crate) match_type: MatchType,
    pub(crate) priority: i32,
    pub(crate) enabled: bool,
    pub(crate) regex: Option<Regex>,
}

impl MatchRule {
    pub(crate) fn new(
        rule_id: Option<String>,
        pattern: String,
        match_type: MatchType,
        priority: i32,
        enabled: bool,
        regex: Option<Regex>,
    ) -> Self {
        Self {
            rule_id,
            normalized_pattern: pattern.to_lowercase(),
            pattern,
            match_type,
            priority,
            enabled,
            regex,
        }
    }
}

#[derive(Debug, Clone)]
pub(crate) struct MatchResult {
    #[allow(dead_code)]
    pub(crate) rule_id: Option<String>,
    pub(crate) pattern: String,
    pub(crate) match_type: MatchType,
    pub(crate) matched_text: String,
    pub(crate) excerpt: String,
}

#[derive(Default, Debug, Clone)]
pub(crate) struct DenyWord {
    contains_rules: Vec<MatchRule>,
    exact_rules: Vec<MatchRule>,
    regex_rules: Vec<MatchRule>,
}

impl DenyWord {
    pub(crate) fn from_rules(mut rules: Vec<MatchRule>) -> Self {
        rules.retain(|rule| rule.enabled && !rule.pattern.trim().is_empty());
        rules.sort_by(|left, right| {
            right
                .priority
                .cmp(&left.priority)
                .then_with(|| right.pattern.len().cmp(&left.pattern.len()))
        });

        let mut deny_word = DenyWord::default();
        for rule in rules {
            match rule.match_type {
                MatchType::Contains => deny_word.contains_rules.push(rule),
                MatchType::Exact => deny_word.exact_rules.push(rule),
                MatchType::Regex => deny_word.regex_rules.push(rule),
            }
        }
        deny_word
    }

    pub(crate) fn empty() -> Self {
        Self::default()
    }

    pub(crate) fn system() -> Self {
        Self::from_system_text(
            Asset::get("sensitive_word_dict.txt")
                .and_then(|file| {
                    std::str::from_utf8(file.data.as_ref())
                        .ok()
                        .map(str::to_string)
                })
                .unwrap_or_default()
                .as_str(),
        )
    }

    #[allow(dead_code)]
    pub(crate) fn from_text(text: &str) -> Self {
        let mut rules = Vec::new();
        for line in text.lines() {
            let pattern = line.trim();
            if pattern.is_empty() {
                continue;
            }
            let regex = Regex::new(&format!("(?i:{})", regex_escape(pattern))).ok();
            rules.push(MatchRule::new(
                None,
                pattern.to_string(),
                MatchType::Contains,
                0,
                true,
                regex,
            ));
        }
        Self::from_rules(rules)
    }

    pub(crate) fn from_system_text(text: &str) -> Self {
        let mut rules = Vec::new();
        for line in text.lines() {
            let pattern = line.trim();
            if pattern.is_empty() {
                continue;
            }
            let (match_type, regex) = if should_use_exact_system_match(pattern) {
                (MatchType::Exact, None)
            } else {
                (
                    MatchType::Contains,
                    Regex::new(&format!("(?i:{})", regex_escape(pattern))).ok(),
                )
            };
            rules.push(MatchRule::new(
                None,
                pattern.to_string(),
                match_type,
                0,
                true,
                regex,
            ));
        }
        Self::from_rules(rules)
    }

    pub(crate) fn append(&mut self, other: Self) {
        self.contains_rules.extend(other.contains_rules);
        self.exact_rules.extend(other.exact_rules);
        self.regex_rules.extend(other.regex_rules);
        self.contains_rules.sort_by(|left, right| {
            right
                .priority
                .cmp(&left.priority)
                .then_with(|| right.pattern.len().cmp(&left.pattern.len()))
        });
        self.exact_rules
            .sort_by(|left, right| right.priority.cmp(&left.priority));
        self.regex_rules
            .sort_by(|left, right| right.priority.cmp(&left.priority));
    }

    pub(crate) fn check(&self, message: &str) -> Option<MatchResult> {
        if message.is_empty() {
            return None;
        }

        let normalized_message = message.to_lowercase();
        for rule in &self.contains_rules {
            if !normalized_message.contains(&rule.normalized_pattern) {
                continue;
            }
            if let Some(regex) = &rule.regex {
                if let Ok(Some(matched)) = regex.find(message) {
                    return Some(build_match_result(
                        rule,
                        message,
                        matched.start(),
                        matched.end(),
                    ));
                }
            }
            return Some(build_match_result(rule, message, 0, message.len()));
        }

        for rule in &self.exact_rules {
            if normalized_message == rule.normalized_pattern {
                return Some(build_match_result(rule, message, 0, message.len()));
            }
        }

        for rule in &self.regex_rules {
            if let Some(regex) = &rule.regex {
                if let Ok(Some(matched)) = regex.find(message) {
                    return Some(build_match_result(
                        rule,
                        message,
                        matched.start(),
                        matched.end(),
                    ));
                }
            }
        }

        None
    }
}

fn build_match_result(rule: &MatchRule, message: &str, start: usize, end: usize) -> MatchResult {
    MatchResult {
        rule_id: rule.rule_id.clone(),
        pattern: rule.pattern.clone(),
        match_type: rule.match_type,
        matched_text: message
            .get(start..end)
            .map(|s| s.to_string())
            .unwrap_or_else(|| rule.pattern.clone()),
        excerpt: build_excerpt(message, start, end),
    }
}

fn build_excerpt(message: &str, start: usize, end: usize) -> String {
    let max_chars = 24usize;
    let mut char_positions = message
        .char_indices()
        .map(|(idx, _)| idx)
        .collect::<Vec<_>>();
    char_positions.push(message.len());

    let start_char_idx = char_positions
        .iter()
        .position(|idx| *idx >= start)
        .unwrap_or(0);
    let end_char_idx = char_positions
        .iter()
        .position(|idx| *idx >= end)
        .unwrap_or(char_positions.len().saturating_sub(1));

    let excerpt_start_char = start_char_idx.saturating_sub(max_chars);
    let excerpt_end_char = (end_char_idx + max_chars).min(char_positions.len().saturating_sub(1));
    let excerpt_start = char_positions.get(excerpt_start_char).copied().unwrap_or(0);
    let excerpt_end = char_positions
        .get(excerpt_end_char)
        .copied()
        .unwrap_or(message.len());

    let mut excerpt = String::new();
    if excerpt_start > 0 {
        excerpt.push_str("...");
    }
    excerpt.push_str(message.get(excerpt_start..excerpt_end).unwrap_or(message));
    if excerpt_end < message.len() {
        excerpt.push_str("...");
    }
    excerpt
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

fn should_use_exact_system_match(pattern: &str) -> bool {
    let normalized = pattern.trim();
    let char_count = normalized.chars().count();
    char_count <= 3 && normalized.chars().all(is_cjk_char)
}

fn is_cjk_char(ch: char) -> bool {
    matches!(
        ch as u32,
        0x3400..=0x4DBF
            | 0x4E00..=0x9FFF
            | 0xF900..=0xFAFF
            | 0x20000..=0x2A6DF
            | 0x2A700..=0x2B73F
            | 0x2B740..=0x2B81F
            | 0x2B820..=0x2CEAF
            | 0x2CEB0..=0x2EBEF
            | 0x30000..=0x3134F
    )
}

#[cfg(test)]
mod tests {
    use super::DenyWord;

    #[test]
    fn short_cjk_system_terms_use_exact_matching() {
        let deny_word = DenyWord::from_system_text("天安门\n天安门事件");

        assert!(deny_word.check("天安门").is_some());
        assert!(deny_word.check("天安门的景点").is_none());
        assert!(deny_word.check("请介绍天安门事件").is_some());
    }
}
