# Sources and generated data

The CIDR lists published by this project combine data from multiple upstream projects and public datasets. These generated lists are derived data and may be subject to the licenses and terms that apply to their upstream sources.

The `ChinaRoute` group includes [misakaio/chnroutes2](https://github.com/misakaio/chnroutes2), an aggregate based on multiple BGP feeds. The upstream project updates its routes hourly and licenses its data under [CC BY-SA](https://github.com/misakaio/chnroutes2#license).

Licensing for the project code is separate from licensing for the generated data. Any license covering the code does not replace, override, or grant rights under the licenses or terms of the upstream data. No single blanket license is asserted for the generated CIDR data.

## Source inventory

The configured source definitions live under [`sources/CN/`](sources/CN/). The [Data sources section in README](README.md#data-sources) lists the upstream projects and URLs currently configured. The generated [`metadata.json`](https://rules.mewrix.com/metadata.json) records the source URLs, response SHA-256 values, and per-source prefix counts for each build. It does not specify or replace the applicable upstream licenses or terms.

Upstream sources and their terms may change independently. Users of the generated data should consult the applicable upstream licenses and terms, including any attribution requirements.
