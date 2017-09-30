# gooder-reads

iptables command:

	iptables -t nat -A PREROUTING -p tcp --dport 443 -j DNAT --to-destination :8443
