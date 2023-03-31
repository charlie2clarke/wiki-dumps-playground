[wikidata]: mediawiki/wikidata-testdata-all.json "mediawiki/wikidata-testdata-all.json"
[osmfilter]: wikiall-output.txt "wikiall-output.txt"
[enterprise]: enterprisedumps/enterprise.go "enterprisedumps/enterprise.go"
[mediawiki]: mediawiki/mediawiki.go "mediawiki/mediawiki.go"
[wikiparse]: wikiparse/wikiparse.go "wikiparse/wikiparse.go"
[osm]: osm/osm.go "osm/osm.go"

[wikidata] Shows a snippet of the JSON dumps format of wikidata. Using the key "sitelinks" you can find the corresponding Wikipedia article in every available language.

[osmfilter] Shows a snippet of the output of the osmfilter tool when applied to the Yukon Canada OSM. This used the following filter to obtain all Wiki data:
```
$ osmfilter yukon-latest.osm --keep="wikidata*=* wikipedia*=*" --keep-tags="all wikidata*=* wikipedia*=*" --ignore-dependencies -o=wikiall-output.txt
```

[enterprise] Includes basic parsing of the enterprise Wikimedia HTML dumps. This data is stored in ndjson format.

[mediawiki] Includes basic parsing of the Wikidata JSON dumps, as well as how the latest dump can be monitored.

[wikiparse] Includes basic parsing of the Wikipedia mulistream XML dumps using the index file.

[osm] Shows extraction of Wiki tags from an osm/o5m file.
