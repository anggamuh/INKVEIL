package main
import("testing";"context")
func TestJellyfinRequiresExplicitConfig(t *testing.T){t.Setenv("INKVEIL_JELLYFIN_PERMISSION","no");if jfEnabled(){t.Fatal("unexpected enable")};if _,ok:=jfFind(context.Background(),"jf-../private",false);ok{t.Fatal("invalid id accepted")}}
