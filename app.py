import autogen.io.base as aio
import streamlit as st
from autogen import ConversableAgent, GroupChatManager


class MyIOStream(aio.IOStream):
    def print(self, *objects: Any, sep: str = " ", end: str = "\n", flush: bool = False) -> None:
        pass

    def input(self, prompt: str = "", *, password: bool = False) -> str:
        return "Hello, World!"


# 创建一个输出流
output_stream = aio.OutputStream()

# 设置默认输出流
aio.IOStream.set_default(MyIOStream())

# 创建代理
agent = ConversableAgent("User", human_input_mode="NEVER")
manager = GroupChatManager(agents=[agent], llm_config={})

# Streamlit 应用
st.title("Autogen 聊天应用")

# 用户输入框
user_input = st.text_input("输入消息")

if st.button("发送"):
    # 发送消息
    reply = agent.initiate_chats(manager, chat_queue=[{"content": user_input, "name": "User"}])

    # 显示回复
    for message in reply:
        st.write(f"{message['name']}: {message['content']}")

# 显示输出流内容
st.subheader("输出流内容")
with st.expander("展开输出流内容"):
    st.write(output_stream.read())
