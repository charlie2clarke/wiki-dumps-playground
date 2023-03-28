[wikidata]: wikidata-testdata-all.json "wikidata-testdata-all.json"
[osmfilter]: wikiall-output.txt "wikiall-output.txt"

[wikidata] shows a snippet of the JSON dumps format of wikidata. Using the key "sitelinks" you can find the corresponding Wikipedia article in every available language.

[osmfilter] shows a snippet of the output of the osmfilter tool when applied to the Yukon Canada OSM. This used the following filter to obtain all Wiki data:
```
$ osmfilter yukon-latest.osm --keep="wikidata*=* wikipedia*=*" --keep-tags="all wikidata*=* wikipedia*=*" --ignore-dependencies -o=wikiall-output.txt
```