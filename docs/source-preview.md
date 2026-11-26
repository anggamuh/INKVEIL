# Offline source preview

Save manga detail, chapter list and chapter detail JSON separately. Existing command remains supported:

`go run . preview-source manga.json chapters.json chapter.json`

For multiple saved chapter list pages, pass them in capture order:

`go run . preview-source manga.json chapters1.json chapters2.json chapter.json`

Up to 50 list pages and 10,000 chapter entries. Rejects duplicate chapter IDs, repeated or missing continuation cursors, empty pages, and pages following a final page. HasNext reflects the last supplied page. ListPages reports number of supplied pages. Saved responses alone cannot prove continuity or completeness, even if the final page reports HasNext=false. No network calls, database writes or image downloads. Use authorized ZIP packages for reader tests.
