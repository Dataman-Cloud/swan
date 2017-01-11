### High Avaliablity

`Swan`设计之初考虑到了高可用目标，
生产环境推荐部署多个Manager，多个Manager内部通过Raft协议做Leader选举，只要保证N/2+1个节点正常运行就不会引起数据丢失。
到Follower上的请求都会被转发到Leader上。
