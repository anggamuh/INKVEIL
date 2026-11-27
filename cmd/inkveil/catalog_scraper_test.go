package main
import("testing";"net/url")
func TestScrapeStructured(t *testing.T){u,_:=url.Parse("https://example.org/catalog");body:=`<script type="application/ld+json">{"@type":"ItemList","itemListElement":[{"item":{"@type":"ComicSeries","name":"Example","url":"/comic/example","author":{"name":"Author"}}},{"item":{"@type":"ComicSeries","name":"Wrong","url":"https://other.example/comic"}}]}</script>`;rows:=scrapeStructured(body,u);if len(rows)!=1||rows[0].Title!="Example"||rows[0].Href!="https://example.org/comic/example"||rows[0].Author!="Author"{t.Fatal(rows)}}
func TestScrapeNoMetadata(t *testing.T){u,_:=url.Parse("https://example.org");if len(scrapeStructured("<html>Browser verification</html>",u))!=0{t.Fatal("unexpected record")}}
