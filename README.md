# cn-service-cidrs

Ready-to-use IPv4 and IPv6 CIDR lists for China, built from community-maintained rule lists and public routing datasets. Entries are combined, deduplicated, and refreshed daily when the upstream data changes.

Non-canonical CIDRs are rejected rather than automatically masked, to avoid silently expanding malformed upstream prefixes.

## Downloads

| List | Download |
| --- | --- |
| Combined CN service & routing prefixes | [CN.txt](https://rules.mewrix.com/tables/CN.txt) |
| Combined CN IPv4 prefixes | [CN-ipv4.txt](https://rules.mewrix.com/tables/CN-ipv4.txt) |
| Combined CN IPv6 prefixes | [CN-ipv6.txt](https://rules.mewrix.com/tables/CN-ipv6.txt) |
| China routing / GeoIP-oriented aggregate | [ChinaRoute.txt](https://rules.mewrix.com/tables/CN/ChinaRoute.txt) |

Browse the [complete collection of generated lists](https://rules.mewrix.com/).

`CN.txt` is the address-set union of every configured source group. Duplicate and overlapping ranges are removed, and adjacent CIDRs are combined whenever they exactly cover a parent prefix. The resulting list preserves the covered addresses but not the source or service category for each address. Use the individual source-group lists when that distinction matters.

The generated `metadata.json` records each configured source's group, type, URL, SHA-256 of the fetched response body, IPv4/IPv6 prefix counts after minimization, rejected CIDR entries, and its exclusive IPv4/IPv6 address counts. An address is exclusive to a source when that source covers it and no other configured source does. Addresses covered by multiple sources are not assigned to any source, so these counts sum to the addresses covered by exactly one source and may be less than the final union. Per-source prefix counts are diagnostic and are not additive because sources can overlap. Invalid CIDRs in plain CIDR sources and non-canonical CIDRs are rejected and counted. Malformed CIDRs in Clash rules are treated as build errors. Its `content_hash` is a deterministic hash of the generated `.txt` tables and is computed during the Go build.

Release notes compare aggregate prefix counts and report source additions, removals, and content or count changes.

## Licensing

The [MIT License](LICENSE) applies only to the project code; it does not apply to generated CIDR data. The generated lists derive from multiple upstream datasets and may be subject to their respective terms. See [SOURCES.md](SOURCES.md) and the source list below for information about the data sources.

## Data sources

### `country-csv`

Country CSV sources contain `ip_range_start`, `ip_range_end`, and `country_code` columns. The header is optional; headerless files use that column order. The parser keeps the configured country code and converts each inclusive IP range into CIDR prefixes.

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
- [misakaio/chnroutes2 IPv4](https://raw.githubusercontent.com/misakaio/chnroutes2/master/chnroutes.txt) ([repository](https://github.com/misakaio/chnroutes2))

#### IP database oriented

- [sapics/ip-location-db server-country IPv4](https://github.com/sapics/ip-location-db/releases/download/latest/server-country-ipv4.csv) and [IPv6](https://github.com/sapics/ip-location-db/releases/download/latest/server-country-ipv6.csv) (country code `CN`)
- [clang.cn all_cn_ipv46.txt](https://ispip.clang.cn/all_cn_ipv46.txt) ([source site](https://ispip.clang.cn))
- [metowolf/iplist China](https://metowolf.github.io/iplist/data/special/china.txt) ([repository](https://github.com/metowolf/iplist))

#### Supplemental

- [17mon/china_ip_list](https://raw.githubusercontent.com/17mon/china_ip_list/master/china_ip_list.txt) ([repository](https://github.com/17mon/china_ip_list))
- [Loyalsoldier/geoip CN](https://raw.githubusercontent.com/Loyalsoldier/geoip/release/text/cn.txt) ([repository](https://github.com/Loyalsoldier/geoip))
- [IPdeny China IPv4 aggregated zone](https://www.ipdeny.com/ipblocks/data/aggregated/cn-aggregated.zone)
- [IPdeny China IPv6 aggregated zone](https://www.ipdeny.com/ipv6/ipaddresses/aggregated/cn-aggregated.zone)
