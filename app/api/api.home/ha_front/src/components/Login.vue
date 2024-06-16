<template>
  <div class="login-container">
    <div class="logo">
      <img src="@/assets/logo.jpg" alt="Logo"/>
    </div>
    <h1 class="title">Vines</h1>
    <p class="subtitle">智联一切</p>
    <form @submit.prevent="handleSubmit">
      <div class="form-group">
        <div class="input-group">
          <input type="text" id="phone" v-model="phone" placeholder="请输入手机号" class="phone-input"/>
          <button type="button" @click="getVerificationCode">获取验证码</button>
        </div>
      </div>
      <div class="form-group">
        <div class="code-group">
          <input type="text" id="code" v-model="code" placeholder="请输入验证码" class="code-input"/>
        </div>
      </div>
      <p class="note">未注册手机号验证通过后将自动注册</p>
      <button type="submit" class="login-button">登录</button>
    </form>
    <div class="agreement">
      <input type="checkbox" id="agreement" v-model="agreed"/>
      <label for="agreement">
        我已阅读并同意
        <a href="#" class="link">《用户协议》</a>
        和
        <a href="#" class="link">《隐私政策》</a>
      </label>
    </div>
  </div>
</template>

<script>
import axios from 'axios';

export default {
  data() {
    return {
      phone: '',
      code: '',
      agreed: false,
    };
  },
  methods: {
    async handleSubmit() {
      if (!this.agreed) {
        alert('请先同意用户协议和隐私政策');
        return;
      }
      try {
        const loginRsp = {
          code: 0,
          msg: "",
          data: {
            redirect_url: "",
            access_token: ""
          }
        };
        const response = await axios.post('http://43.139.244.233:10005/home/user/login', {
          phone: this.phone,
          code: this.code,
        });

        Object.assign(loginRsp, response.data);

        if (loginRsp.code === 200) {
          const accessToken = loginRsp.data.access_token;
          localStorage.setItem('access_token', accessToken);
          alert('登录成功');
          // 可以在这里进行后续操作，例如跳转页面
          // 获取重定向 URL
          window.location.href = "http://"+loginRsp.data.redirect_url;
        } else {
          alert('登录失败：' + loginRsp.msg);
        }
      } catch (error) {
        alert('请求失败：' + error.message);
      }
    },
    async getVerificationCode() {
      if (!this.phone) {
        alert('请输入手机号');
        return;
      }
      try {
        const response = await axios.post('http://43.139.244.233:10005/home/user/code', {
          phone: this.phone,
        });
        if (response.data.code === 200) {
          alert(response.data.msg);
        } else {
          alert('获取验证码失败：' + response.data.msg);
        }
      } catch (error) {
        alert('请求失败：' + error.message);
      }
    },
  },
};
</script>

<style scoped>
.login-container {
  max-width: 400px;
  height: 100vh;
  margin: 0 auto;
  padding: 20px;
  text-align: center;
  font-family: Arial, sans-serif;
  background-image: url('@/assets/background.png'); /* 设置背景图 */
  background-size: cover; /* 背景图覆盖整个容器 */
  background-position: center; /* 背景图居中 */
}

.logo img {
  width: 100px;
  margin-bottom: 10px;
}

.title {
  font-size: 24px;
  margin-bottom: 5px;
}

.subtitle {
  font-size: 16px;
  color: #888;
  margin-bottom: 20px;
}

.form-group {
  margin-bottom: 15px;
}

.input-group {
  display: flex;
  align-items: center;
}

.phone-input {
  flex: 2;
  padding: 8px;
  border: 1px solid #ccc;
  border-radius: 4px;
  margin-right: 10px;
}

.code-group {
  display: flex;
  justify-content: flex-start; /* 左对齐 */
}

.code-input {
  width: 30%; /* 设置验证码输入框宽度为手机号输入框的一半 */
  padding: 8px;
  border: 1px solid #ccc;
  border-radius: 4px;
}

.input-group button {
  flex: 1;
  padding: 8px 12px;
  background-color: #007bff;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
}

.input-group button:hover {
  background-color: #0056b3;
}

.note {
  font-size: 12px;
  color: #888;
  margin-bottom: 20px;
}

.login-button {
  width: 100%;
  padding: 10px 0;
  background-color: #ff4d4f;
  color: white;
  border: none;
  border-radius: 4px;
  font-size: 16px;
  cursor: pointer;
}

.login-button:hover {
  background-color: #d43f3a;
}

.agreement {
  margin-top: 20px;
  font-size: 14px;
  text-align: left;
}

.agreement input {
  margin-right: 5px;
}

.agreement .link {
  color: #007bff;
  text-decoration: none;
}

.agreement .link:hover {
  text-decoration: underline;
}
</style>
