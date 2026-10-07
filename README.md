# cn-service-cidrs

Ready-to-use IPv4 and IPv6 CIDR lists plus domain rules for China, built from community-maintained rule lists and public routing datasets. Entries are combined, deduplicated, and refreshed daily when the upstream data changes.

Non-canonical CIDRs are rejected rather than automatically masked, to avoid silently expanding malformed upstream prefixes.

## Downloads

| List | Download |
| --- | --- |
| Combined CN service & routing prefixes | [CN.txt](https://rules.mewrix.com/tables/CN.txt) |
| Combined CN IPv4 prefixes | [CN-ipv4.txt](https://rules.mewrix.com/tables/CN-ipv4.txt) |
| Combined CN IPv6 prefixes | [CN-ipv6.txt](https://rules.mewrix.com/tables/CN-ipv6.txt) |
| Combined CN domain rules | [CN-domain.txt](https://rules.mewrix.com/tables/CN-domain.txt) |
| AdGuard Home domain upstreams | [aghome-upstream.txt](https://rules.mewrix.com/tables/aghome-upstream.txt) |
| Technitium Advanced Forwarding domain upstreams | [technitium-upstream.txt](https://rules.mewrix.com/tables/technitium-upstream.txt) |
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

### `domain`

Domain sources support plain domain lists, Clash-style rules, v2fly domain-list-community exports, dnsmasq `server=/domain/DNS` entries, and the Surge ChinaMax domain list. Rules are emitted with the project prefixes `domain:`, `full:`, `regexp:`, and `keyword:`; unsupported non-domain rules are ignored. v2fly's `domain`, `full`, `regexp`, and `keyword` types are preserved, while list attributes are discarded. dnsmasq entries become `domain:` suffix rules. In Surge ChinaMax, a leading dot becomes a `domain:` suffix rule and an unprefixed hostname becomes a `full:` exact rule. The mapped type and rule value are kept in separate `*-domain.txt` files, outside all CIDR lists and aggregates. Duplicate rules are removed. A broader `domain:` suffix also subsumes narrower `domain:` suffixes and covered `full:` rules; `regexp:` and `keyword:` rules are not merged. Blank lines and lines beginning with `#` are ignored. For example, `DOMAIN-SUFFIX,bilibili.com` becomes `domain:bilibili.com`, `DOMAIN,example.cn` becomes `full:example.cn`, and `DOMAIN-REGEX,^foo[0-9]+\.example\.cn$` becomes `regexp:^foo[0-9]+\.example\.cn$`.

The build also writes `output/tables/aghome-upstream.txt` using AdGuard Home's domain-specific upstream syntax. It includes minimized `domain:` suffix rules; `full:`, `regexp:`, and `keyword:` rules are excluded because this upstream syntax cannot preserve their matching semantics. Suffix domains are converted to lowercase IDNA ASCII/Punycode and stripped of a trailing dot because AdGuard Home's upstream configuration does not accept Unicode IDNs. Domains sharing the same upstreams are grouped into rules of at most 100 domains and 4 KiB each. For example, one output line can be `[/example.com/example.cn/]https://dns.alidns.com/dns-query https://doh.pub/dns-query`. Add the file's contents to AdGuard Home's upstream DNS configuration. The `AGHOME_UPSTREAM_DNS` GitHub Actions variable accepts multiple upstreams separated by spaces and is shared by the Pages and CI workflows. Its default is `https://dns.alidns.com/dns-query https://doh.pub/dns-query` (Alibaba Cloud Public DNS and DNSPod Public DNS). Set that repository variable to change the upstreams. Multiple upstreams in a domain-specific entry require AdGuard Home v0.107.41 or later.

The build also writes `output/tables/technitium-upstream.txt` for Technitium DNS Server's Advanced Forwarding. It uses the same minimized, IDNA ASCII `domain:` suffixes and grouping as the AdGuard Home file, with a bootstrap IP in parentheses after each DoH URL so Technitium can reach the resolver without first resolving its hostname. For example: `[/example.com/example.cn/]https://dns.alidns.com/dns-query (223.5.5.5)`. The default is Alibaba Cloud Public DNS; set the `TECHNITIUM_UPSTREAM_DNS` GitHub Actions variable to change the configured upstream and bootstrap IP. This file currently contains domain forwarding rules; it can be extended with IP prefix rules separately.

```yaml
name: BiliBiliDomains
sources:
  - type: domain
    value: https://example.com/bilibili-domains.txt
```

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
- [ChinaMax_Domain](https://raw.githubusercontent.com/blackmatrix7/ios_rule_script/master/rule/Surge/ChinaMax/ChinaMax_Domain.list)

### Domain sources

- [v2fly/domain-list-community `cn` export](https://raw.githubusercontent.com/v2fly/domain-list-community/release/cn.txt), which combines `geolocation-cn` and `tld-cn`
- [felixonmars/dnsmasq-china-list accelerated domains](https://raw.githubusercontent.com/felixonmars/dnsmasq-china-list/master/accelerated-domains.china.conf)

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
