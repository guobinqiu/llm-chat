# LLM Chat

- chat-once 一次性对话
- chat-loop 多轮对话
  - idiot 没有记忆功能
  - smart 有记忆功能但是累计的历史对话会造成token超限和网络传输的开销
  - smarter 有记忆功能同时优化了性能
    - 上下文裁剪（保留最近对话）
    - 摘要长对话内容 （压缩旧消息）