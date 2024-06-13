如果在 Docker 容器中运行 apt-get 命令时出现 command not found 错误，这意味着基础镜像没有安装 apt-get。Home Assistant 的 Docker 镜像可能基于 Alpine Linux，它使用 apk 包管理器而不是 apt-get。

以下是基于 Alpine Linux 的步骤来安装所需的依赖项：

1. 进入 Home Assistant 容器
   首先，进入正在运行的 Home Assistant 容器：

bash
复制
docker exec -it home-assistant /bin/bash
2. 安装依赖项
   在容器内部，运行以下命令以安装蓝牙相关的依赖项和 PyBluez：

sh
复制
apk update
apk add --no-cache bluez bluez-dev py3-pip
pip install PyBluez==0.22
3. 验证安装
   确保 PyBluez 已正确安装：

sh
复制
pip show PyBluez
4. 退出容器并重启
   退出容器并重启 Home Assistant 以应用更改：

sh
复制
exit
docker restart home-assistant
5. 检查配置文件
   确保你的 configuration.yaml 文件中 bluetooth_tracker 配置正确：

yaml
复制
device_tracker:
- platform: bluetooth_tracker
  track_new_devices: true
6. 检查日志
   查看 Home Assistant 的日志文件确认 bluetooth_tracker 是否正常工作：

sh
复制
docker logs home-assistant
7. 持久化依赖（可选）
   每次重启容器时，这些安装的依赖项可能会丢失。为了解决这个问题，你可以创建一个自定义的 Docker 镜像，其中包含所有必要的依赖项。

创建自定义 Dockerfile
创建一个名为 Dockerfile 的文件，并添加以下内容：

dockerfile
复制
FROM homeassistant/home-assistant:latest

RUN apk update && \
apk add --no-cache bluez bluez-dev py3-pip && \
pip install PyBluez==0.22
构建自定义镜像
运行以下命令构建自定义的 Home Assistant Docker 镜像：

sh
复制
docker build -t my-home-assistant .
运行自定义镜像
使用自定义镜像运行 Home Assistant 容器：

sh
复制
docker run -d \
--name home-assistant \
--restart unless-stopped \
--net=host \
--privileged \
-v /path/to/your/config:/config \
-v /run/dbus:/run/dbus \
-v /var/run/dbus:/var/run/dbus \
--device /dev/ttyUSB0 \
--device /dev/ttyACM0 \
--device /dev/hci0 \
my-home-assistant
通过这些步骤，你应该能够在 Docker 容器中成功启用 bluetooth_tracker，并解决 PyBluez 依赖问题。如果遇到其他问题，请提供更多细节以便进一步诊断。