

#### proxy statistics
`GET` `/proxy/stats`

```json
{
  "global": {
    "rx_bytes": 961963,
    "tx_bytes": 10165195,
    "requests": 14730,
    "fails": 2785,
    "rx_rate": 5280,
    "tx_rate": 58293,
    "requests_rate": 112,
    "fails_rate": 44,
    "uptime": "4m17.745790884s"
  },
  "app": {
    "aaa-default-zgz-datamanmesos": {
      "1-aaa-default-zgz-datamanmesos": {
        "active_clients": 1,
        "rx_bytes": 618524,
        "tx_bytes": 5211524,
        "requests": 6124,
        "rx_rate": 2979,
        "tx_rate": 25104,
        "requests_rate": 29,
        "uptime": "3m17.11239229s"
      }
    },
    "ccc-default-zgz-datamanmesos": {
      "0-ccc-default-zgz-datamanmesos": {
        "active_clients": 0,
        "rx_bytes": 343439,
        "tx_bytes": 4953671,
        "requests": 5821,
        "rx_rate": 2301,
        "tx_rate": 33189,
        "requests_rate": 39,
        "uptime": "2m31.447905211s"
      }
    }
  }
}
```
