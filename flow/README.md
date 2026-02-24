# Flow

本地内存流式处理框架

## Message

Accept() 是重点, 通过访问者模式解耦流式框架和业务逻辑, 开发者可以针对每个消息编写逻辑.

无论多复杂的场景都能通过 Accept() 和 Visitor 的设计来适配:

- 消息订阅分发
- 消息转换

## Endpoint

.Activate() 是开发者提供的初始化某个端点的回调, 框架不会假设这个回调的实现是可重入的, 而是由框架保证每个 Activate() 只会调用一次.

flow.go 提供的 builder 在构造 flow 时, 会显式的允许开发者指定哪些 Endpoint 是 Eager, 哪些是 Lazy. 其实就是通过不同的 builder method 告诉 builder 哪些 Endpoint 立刻执行一次 Activate, 哪些 Endpoint 可以先不初始化, 等等看有没有某个 Message 出现时要求激活某些 Endpoint.

至于 .Activate() 实现本身, 尤其是 Eager 的 Endpoint, 可能在框架外就初始化了, 所以这里 return nil 就行, 主要是框架本身得"运行一次并登记哪些 Endpoint 是活跃的"

## Source

只需要提供一个类似迭代器的 Next() 函数, 比如 stream.Recv() 或者是一个数组迭代器等.

框架本身会拉起 routine, 将 Pull 模式的 Source 转变成 Push 模式来适配框架逻辑

## Sink

消息最终的归属, 到了 Sink 后就会被排出去, 只需要提供消费的逻辑 Drain(), 比如: stream.Send()

## Duplex

Duplex 的概念其实可以理解为那些全双工通信的端点, 它们既产生数据, 也消费数据, 但它们往往公用一个初始化过程, 即一个 Activate() 就会同时激活 Source 和 Sink

## 并发控制

Source 和 Sink 是串行执行的, 框架内部有一个没有导出的类 Actor, 这个 Actor 是并发执行 message.Accept().

当 message.DrainTo() 返回非 None 时, 会由对应的 Sink 串行执行 message.Accept().

综上可以理解为, message.DrainTo() 为 None 时, 消息处理会在并发环境, 非 None 时会在 Sink 串行执行环境处理消息

## Implement

实现上, Endpoint, Source, Sink 基本都会有一个实现类去对应, 它们包装开发者实现的 interface 实例, 通过回调开发者的实现, 配合流式框架完成流式处理的任务

## 工具函数

框架为了足够的灵活性考虑, 有很多重要但不是全场景通用的逻辑是没有实现的, 比如顺序消费等, 这些逻辑如果全部让开发者自行在 Visitor 实现, 必然是很痛苦的事情.

框架可以提供大量的工具函数/工具类, 尽可能让开发者挑选他们需要的工具函数, 包装他们对框架的 interfaces 的实现就能完成许多通用能力的实现

## Testing

测试时, 可以使用 Mockery 生成的 interfaces 实现来模拟开发者的回调
