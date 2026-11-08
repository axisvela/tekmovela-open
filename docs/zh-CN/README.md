# TEKMOVELA 文档（中文）

本文档以英文为主；此页面提供中文导航链接�?
> 英文文档是权威来源，中文文档可能滞后。如有出入，以英文为准�?
## 入门

- [产品简介（中文）](../README.zh-CN.md) �?TEKMOVELA 是什么、解决什么问�?- [Architecture overview（英文）](../architecture/overview.md)
- [API reference（英文）](../api.md)
- [OSS / EE boundary（英文）](../editions.md)
- [Roadmap（英文）](../roadmap.md)

## 产品

TEKMOVELA �?AxisVela �?*自主系统工程保障控制平面**。它将意图与验收标准转化为验证合约与受治理的测试装置，执行可复现的自主系统闭环测试，�?trace 关联证据诊断失败，把关发布，并从生产行为持续触发再验证�?

## 边缘与学习边界

TEKMOVELA 拥有独立保证，可通过受治理 harness 执行测试工作负载，但不拥有生产
执行。AKROVELA receipt 与 DIDACTOVELA 训练指标只是生产者输入，不是独立 pass。
DIDACTOVELA 拥有数据集/标签/训练/候选谱系；NOETIVELA 拥有生产模型、端点、
路由与激活。导入候选既不激活端点，也不满足独立保证。

AKROVELA 与 DIDACTOVELA 仅为未发布 `0.1.0` bootstrap，没有已实现的领域集成。
未来适配器须在依赖实现前审查版本化公开契约，绑定产物/claim 版本与 digest、
租户/episode 及适用身份/节点 epoch。负向测试须覆盖过期 epoch、伪造 receipt、
跨租户/episode 替换、训练/评估泄漏和生产者自签 pass；不暗示工作集成或新 API。

## 链接

- [English homepage](../../README.md)
- [Contracts（合约）](../../contracts/)
- [Go SDK](../../sdk/go/)
- [Python SDK](../../sdk/python/)
- [Examples（示例）](../../examples/)

## 许可

- `tekmovela-open`（本仓库）�?Apache License 2.0
