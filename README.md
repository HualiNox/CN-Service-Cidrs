# cn-service-cidrs

Generate IPv4 and IPv6 CIDR tables from rule lists and plain CIDR sources. Source definitions are stored under `sources/`; `go run .` writes the tables to `output/tables/` and separately builds the static website in `_site/`. GitHub Pages publishes only `_site/`.

The GitHub Actions workflow rebuilds the site on pushes to `main`, daily, and when manually dispatched. To enable deployment, set **Settings → Pages → Build and deployment → Source** to **GitHub Actions**.

The generated site includes an index with update time and prefix counts, copies of downloadable tables under `_site/tables/`, and machine-readable counts and links in `_site/metadata.json`.

## Data sources

### `clash-list`

#### [LM-Firefly/Rules](https://github.com/LM-Firefly/Rules)

- [360CCC](https://raw.githubusercontent.com/LM-Firefly/Rules/refs/heads/master/CCC-CN/360CCC.list)
- [BaiduCCC](https://raw.githubusercontent.com/LM-Firefly/Rules/refs/heads/master/CCC-CN/BaiduCCC.list)
- [HuaweiCCC](https://raw.githubusercontent.com/LM-Firefly/Rules/refs/heads/master/CCC-CN/HuaweiCCC.list)
- [JingdongCCC](https://raw.githubusercontent.com/LM-Firefly/Rules/refs/heads/master/CCC-CN/JingdongCCC.list)
- [KingsoftCCC](https://raw.githubusercontent.com/LM-Firefly/Rules/refs/heads/master/CCC-CN/KingsoftCCC.list)
- [NeteaseCCC](https://raw.githubusercontent.com/LM-Firefly/Rules/refs/heads/master/CCC-CN/NeteaseCCC.list)
- [TencentCCC](https://raw.githubusercontent.com/LM-Firefly/Rules/refs/heads/master/CCC-CN/TencentCCC.list)
- [Alibaba](https://raw.githubusercontent.com/LM-Firefly/Rules/refs/heads/master/Domestic-Services/Alibaba.list)
- [BiliBili](https://raw.githubusercontent.com/LM-Firefly/Rules/refs/heads/master/Domestic-Services/BiliBili.list)
- [ByteDance](https://raw.githubusercontent.com/LM-Firefly/Rules/refs/heads/master/Domestic-Services/ByteDance.list)
- [ChinaMobile](https://raw.githubusercontent.com/LM-Firefly/Rules/refs/heads/master/Domestic-Services/ChinaMobile.list)
- [ChinaUnicom](https://raw.githubusercontent.com/LM-Firefly/Rules/refs/heads/master/Domestic-Services/ChinaUnicom.list)
- [iQIYI](https://raw.githubusercontent.com/LM-Firefly/Rules/refs/heads/master/Domestic-Services/iQIYI.list)
- [Kugou&Kuwo](https://raw.githubusercontent.com/LM-Firefly/Rules/refs/heads/master/Domestic-Services/Kugou%26Kuwo.list)
- [KuangShi](https://raw.githubusercontent.com/LM-Firefly/Rules/refs/heads/master/Domestic-Services/KuangShi.list)
- [MeiTu](https://raw.githubusercontent.com/LM-Firefly/Rules/refs/heads/master/Domestic-Services/MeiTu.list)
- [Netease](https://raw.githubusercontent.com/LM-Firefly/Rules/refs/heads/master/Domestic-Services/Netease.list)
- [NeteaseMusic](https://raw.githubusercontent.com/LM-Firefly/Rules/refs/heads/master/Domestic-Services/NeteaseMusic.list)
- [Tencent](https://raw.githubusercontent.com/LM-Firefly/Rules/refs/heads/master/Domestic-Services/Tencent.list)
- [Xiaomi](https://raw.githubusercontent.com/LM-Firefly/Rules/refs/heads/master/Domestic-Services/Xiaomi.list)

#### [blackmatrix7/ios_rule_script](https://github.com/blackmatrix7/ios_rule_script)

- [ChinaIPs](https://raw.githubusercontent.com/blackmatrix7/ios_rule_script/master/rule/Clash/ChinaIPs/ChinaIPs.list)

### `cidr`

#### APNIC based — [mayaxcn/china-ip-list](https://github.com/mayaxcn/china-ip-list)

- [IPv4 routes](https://raw.githubusercontent.com/mayaxcn/china-ip-list/master/chnroute.txt)
- [IPv6 routes](https://raw.githubusercontent.com/mayaxcn/china-ip-list/master/chnroute_v6.txt)

#### BGP / operator oriented

- [gaoyifan/china-operator-ip IPv4](https://raw.githubusercontent.com/gaoyifan/china-operator-ip/ip-lists/china.txt) ([repository](https://github.com/gaoyifan/china-operator-ip))
- [gaoyifan/china-operator-ip IPv6](https://raw.githubusercontent.com/gaoyifan/china-operator-ip/ip-lists/china6.txt) ([repository](https://github.com/gaoyifan/china-operator-ip))

#### IP database oriented

- [clang.cn all_cn_ipv46.txt](https://ispip.clang.cn/all_cn_ipv46.txt) ([source site](https://ispip.clang.cn))
- [metowolf/iplist China](https://metowolf.github.io/iplist/data/special/china.txt) ([repository](https://github.com/metowolf/iplist))

#### Supplemental

- [17mon/china_ip_list](https://raw.githubusercontent.com/17mon/china_ip_list/master/china_ip_list.txt) ([repository](https://github.com/17mon/china_ip_list))
- [Loyalsoldier/geoip CN](https://raw.githubusercontent.com/Loyalsoldier/geoip/release/text/cn.txt) ([repository](https://github.com/Loyalsoldier/geoip))
