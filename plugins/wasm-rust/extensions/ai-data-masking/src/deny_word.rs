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

use std::collections::HashMap;

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
    contains_matcher: Option<TokenDfaMatcher>,
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
        deny_word.contains_matcher = TokenDfaMatcher::build(&deny_word.contains_rules);
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
            rules.push(MatchRule::new(
                None,
                pattern.to_string(),
                MatchType::Contains,
                0,
                true,
                None,
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
                (MatchType::Contains, None)
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
        self.contains_matcher = TokenDfaMatcher::build(&self.contains_rules);
    }

    pub(crate) fn check(&self, message: &str) -> Option<MatchResult> {
        if message.is_empty() {
            return None;
        }

        if let Some(matcher) = &self.contains_matcher {
            if let Some(match_result) = matcher.find(message) {
                return Some(match_result);
            }
        }

        let normalized_message = message.to_lowercase();
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

#[derive(Debug, Clone)]
struct TokenSpan {
    normalized: String,
    start: usize,
    end: usize,
}

#[derive(Default, Debug, Clone)]
struct SegmentTrieNode {
    next: HashMap<char, usize>,
    terminal: bool,
}

#[derive(Default, Debug, Clone)]
struct SegmentTrie {
    nodes: Vec<SegmentTrieNode>,
}

impl SegmentTrie {
    fn build(rules: &[MatchRule]) -> Self {
        let mut trie = SegmentTrie {
            nodes: vec![SegmentTrieNode::default()],
        };
        for rule in rules {
            for token in rule
                .pattern
                .split_whitespace()
                .map(str::trim)
                .filter(|item| !item.is_empty() && item.chars().all(is_cjk_char))
            {
                trie.insert(token.to_lowercase());
            }
        }
        trie
    }

    fn insert(&mut self, token: String) {
        let mut node_idx = 0usize;
        for ch in token.chars() {
            let next_idx = if let Some(next_idx) = self.nodes[node_idx].next.get(&ch) {
                *next_idx
            } else {
                let next_idx = self.nodes.len();
                self.nodes.push(SegmentTrieNode::default());
                self.nodes[node_idx].next.insert(ch, next_idx);
                next_idx
            };
            node_idx = next_idx;
        }
        self.nodes[node_idx].terminal = true;
    }

    fn longest_match(&self, chars: &[char], start: usize) -> Option<usize> {
        let mut node_idx = 0usize;
        let mut best_end = None;
        for (offset, ch) in chars.iter().enumerate().skip(start) {
            let Some(next_idx) = self.nodes[node_idx].next.get(ch) else {
                break;
            };
            node_idx = *next_idx;
            if self.nodes[node_idx].terminal {
                best_end = Some(offset + 1);
            }
        }
        best_end
    }
}

#[derive(Debug, Clone)]
struct TokenRule {
    rule: MatchRule,
    tokens: Vec<String>,
}

#[derive(Default, Debug, Clone)]
struct TokenDfaNode {
    next: HashMap<String, usize>,
    outputs: Vec<usize>,
}

#[derive(Debug, Clone)]
struct TokenMatchCandidate {
    rule_index: usize,
    token_start: usize,
    token_end: usize,
}

#[derive(Default, Debug, Clone)]
struct TokenDfaMatcher {
    nodes: Vec<TokenDfaNode>,
    lexicon: SegmentTrie,
    rules: Vec<TokenRule>,
}

impl TokenDfaMatcher {
    fn build(rules: &[MatchRule]) -> Option<Self> {
        if rules.is_empty() {
            return None;
        }
        let lexicon = SegmentTrie::build(rules);
        let mut matcher = TokenDfaMatcher {
            nodes: vec![TokenDfaNode::default()],
            lexicon,
            rules: Vec::new(),
        };
        for rule in rules {
            let tokens = tokenize_message(&rule.pattern, &matcher.lexicon)
                .into_iter()
                .map(|token| token.normalized)
                .collect::<Vec<_>>();
            if tokens.is_empty() {
                continue;
            }
            let rule_index = matcher.rules.len();
            matcher.rules.push(TokenRule {
                rule: rule.clone(),
                tokens: tokens.clone(),
            });
            let mut node_idx = 0usize;
            for token in tokens {
                let next_idx = if let Some(next_idx) = matcher.nodes[node_idx].next.get(&token) {
                    *next_idx
                } else {
                    let next_idx = matcher.nodes.len();
                    matcher.nodes.push(TokenDfaNode::default());
                    matcher.nodes[node_idx].next.insert(token, next_idx);
                    next_idx
                };
                node_idx = next_idx;
            }
            matcher.nodes[node_idx].outputs.push(rule_index);
        }
        Some(matcher)
    }

    fn find(&self, message: &str) -> Option<MatchResult> {
        let tokens = tokenize_message(message, &self.lexicon);
        if tokens.is_empty() {
            return None;
        }
        let mut best: Option<TokenMatchCandidate> = None;
        for start in 0..tokens.len() {
            let mut node_idx = 0usize;
            for end in start..tokens.len() {
                let token = &tokens[end].normalized;
                let Some(next_idx) = self.nodes[node_idx].next.get(token) else {
                    break;
                };
                node_idx = *next_idx;
                for rule_index in &self.nodes[node_idx].outputs {
                    let candidate = TokenMatchCandidate {
                        rule_index: *rule_index,
                        token_start: start,
                        token_end: end + 1,
                    };
                    if prefer_candidate(self, &candidate, best.as_ref()) {
                        best = Some(candidate);
                    }
                }
            }
        }
        let candidate = best?;
        let rule = &self.rules[candidate.rule_index].rule;
        let start = tokens[candidate.token_start].start;
        let end = tokens[candidate.token_end - 1].end;
        Some(build_match_result(rule, message, start, end))
    }
}

fn prefer_candidate(
    matcher: &TokenDfaMatcher,
    candidate: &TokenMatchCandidate,
    current: Option<&TokenMatchCandidate>,
) -> bool {
    let candidate_rule = &matcher.rules[candidate.rule_index];
    let Some(current) = current else {
        return true;
    };
    let current_rule = &matcher.rules[current.rule_index];
    candidate_rule
        .rule
        .priority
        .cmp(&current_rule.rule.priority)
        .then_with(|| candidate_rule.tokens.len().cmp(&current_rule.tokens.len()))
        .then_with(|| candidate_rule.rule.pattern.len().cmp(&current_rule.rule.pattern.len()))
        .then_with(|| current.token_start.cmp(&candidate.token_start))
        .is_gt()
}

fn tokenize_message(message: &str, lexicon: &SegmentTrie) -> Vec<TokenSpan> {
    let mut tokens = Vec::new();
    let mut iter = message.char_indices().peekable();
    while let Some((start_idx, ch)) = iter.peek().copied() {
        if ch.is_whitespace() || is_separator_char(ch) {
            iter.next();
            continue;
        }
        if is_ascii_token_char(ch) {
            let mut end_idx = message.len();
            let mut normalized = String::new();
            while let Some((current_idx, current_ch)) = iter.peek().copied() {
                if !is_ascii_token_char(current_ch) {
                    end_idx = current_idx;
                    break;
                }
                normalized.extend(current_ch.to_lowercase());
                iter.next();
            }
            if normalized.is_empty() {
                continue;
            }
            tokens.push(TokenSpan {
                normalized,
                start: start_idx,
                end: end_idx,
            });
            continue;
        }
        if is_cjk_char(ch) {
            let mut run = Vec::new();
            while let Some((current_idx, current_ch)) = iter.peek().copied() {
                if !is_cjk_char(current_ch) {
                    break;
                }
                run.push((current_idx, current_ch));
                iter.next();
            }
            tokenize_cjk_run(message, &run, lexicon, &mut tokens);
            continue;
        }
        iter.next();
    }
    tokens
}

fn tokenize_cjk_run(
    message: &str,
    run: &[(usize, char)],
    lexicon: &SegmentTrie,
    tokens: &mut Vec<TokenSpan>,
) {
    let chars = run.iter().map(|(_, ch)| ch.to_ascii_lowercase()).collect::<Vec<_>>();
    let mut index = 0usize;
    while index < run.len() {
        let end_index = lexicon.longest_match(&chars, index).unwrap_or(index + 1);
        let start_byte = run[index].0;
        let end_byte = if end_index < run.len() {
            run[end_index].0
        } else {
            message.len()
        };
        let normalized = message[start_byte..end_byte].to_lowercase();
        tokens.push(TokenSpan {
            normalized,
            start: start_byte,
            end: end_byte,
        });
        index = end_index;
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

fn is_ascii_token_char(ch: char) -> bool {
    ch.is_ascii_alphanumeric() || matches!(ch, '_' | '-' | '@')
}

fn is_separator_char(ch: char) -> bool {
    matches!(
        ch,
        ',' | '.'
            | '，'
            | '。'
            | '、'
            | ';'
            | '；'
            | ':'
            | '：'
            | '!'
            | '！'
            | '?'
            | '？'
            | '('
            | ')'
            | '（'
            | '）'
            | '['
            | ']'
            | '{'
            | '}'
            | '<'
            | '>'
            | '"'
            | '\''
            | '`'
            | '/'
            | '\\'
            | '|'
            | '+'
            | '='
            | '*'
            | '#'
            | '&'
            | '^'
            | '%'
            | '$'
            | '~'
            | '\n'
            | '\r'
            | '\t'
    )
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

    #[test]
    fn contains_rules_match_on_segmented_tokens() {
        let deny_word = DenyWord::from_text("南京\nsex");

        assert!(deny_word.check("南京怎么样").is_some());
        assert!(deny_word.check("my sex life").is_some());
        assert!(deny_word.check("sexy style").is_none());
    }
}
