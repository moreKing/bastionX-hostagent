CGO_ENABLED=0 go build -o host-agent

存放服务
/etc/systemd/system/host-agent.service

sudo systemctl daemon-reload

sudo systemctl start host-agent.service

sudo systemctl enable --now host-agent.service