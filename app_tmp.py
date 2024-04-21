import streamlit as st
from autogen import ConversableAgent
import os
import autogen

llm_config = {"model": "gpt-4-turbo-preview", "api_key": "sk-12lth8N6I40ye2cvAd07Cb44Ff84445b918eCa93829eAf7a",
              "base_url": "https://ai-yyds.com/v1"}

job = st.selectbox("选择岗位名称", ('go语言开发工程师', '数据分析师', '保姆'))

if job:
    # 初始化多个代理

    user_proxy = ConversableAgent(
        name="User_proxy",
        system_message="A human admin.",
        human_input_mode="NEVER",
    )

    interviwer_agent = ConversableAgent(
        "interviwer",
        system_message=f"“”你是一个{job}岗位的技术面试官，你将进行一场专业的面试，你需要对面试者进行提问以判断面试者的专业水平，"
                       f"但最好一次只提问一个问题“”",
        llm_config={"config_list": [llm_config]},
        code_execution_config=False,
        function_map=None,
        human_input_mode="NEVER",

    )

    hr_agent = ConversableAgent(
        "hr",
        system_message="你是一个hr，你与interviwer两个人一起对求职者进行面试",
        llm_config={"config_list": [llm_config]},
        code_execution_config=False,
        function_map=None,
        human_input_mode="NEVER",
        is_termination_msg=lambda msg: "今天的面试就到这里" in msg["content"]
    )

    candidate_agent = ConversableAgent(
        "candidate",
        system_message=f"你是一个{job}岗位的求职者，正在参加一场面试以争取拿到offer",
        llm_config={"config_list": [llm_config]},
        code_execution_config=False,
        function_map=None,
        human_input_mode="NEVER",
    )

    groupchat = autogen.GroupChat(agents=[interviwer_agent, hr_agent, candidate_agent], messages=[], max_round=12)
    manager = autogen.GroupChatManager(groupchat=groupchat, llm_config=llm_config)

    st.title("AI群聊应用 - 模拟面试")

    if "messages" not in st.session_state:
        st.session_state["messages"] = []

    chat_container = st.container()

    with chat_container:
        for msg in st.session_state.messages:
            if msg['role'] == 'User':
                st.markdown(f"**{msg['role']}**: {msg['content']}")
            else:
                st.markdown(f"*{msg['role']}*: {msg['content']}")

    if prompt := st.chat_input():
        st.session_state.messages.append({"role": "User", "content": prompt})

        with chat_container:
            st.markdown(f"**User**: {prompt}")

        reply = user_proxy.initiate_chat(manager, message=prompt)

        # 提取角色名称和回答内容
        for message in reply.chat_history:
            if message['role'] == 'user':
                st.session_state.messages.append({"role": message['name'], "content": message['content']})
            elif message['role'] == 'assistant':
                st.session_state.messages.append({"role": "Assistant", "content": message['content']})

        # 显示新的消息
        for msg in st.session_state.messages[-len(reply.chat_history):]:
            with chat_container:
                if msg['role'] == 'User' or msg['role'] == 'Assistant':
                    st.markdown(f"**{msg['role']}**: {msg['content']}")
                else:
                    st.markdown(f"*{msg['role']}*: {msg['content']}")
