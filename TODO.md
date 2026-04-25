# Higress TODO

> 本文件只记录 Higress 子项目自身轻量待办。跨项目专项分析统一索引到根目录 `TASK/projects/higress/`。

## 待补

- [ ] 梳理 `ai-proxy`、`mcp-server`、`model-router` 三类 AI 相关插件的实现边界与配置入口。
- [ ] 梳理 Higress Helm values、CRD、WasmPlugin、Ingress、ConfigMap 之间的真相源关系。
- [ ] 补齐与 `aigateway-console`、`aigateway-portal` 的运行时对象映射说明。
- [ ] 深入分析 `pkg/ingress/translation/` 中 Ingress -> EnvoyFilter 的转换逻辑。
- [ ] 跟踪 Higress 上游 / 子模块版本变化对 AIGateway release 的影响。
