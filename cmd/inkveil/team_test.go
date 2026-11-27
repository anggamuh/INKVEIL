package main
import("testing";"path/filepath")
func TestTeamWorkflow(t *testing.T){
 stages:=[]string{"draft","translation","typesetting","qc","ready","released"}
 for i:=0;i<len(stages)-1;i++{if nextTeamStage(stages[i])!=stages[i+1]{t.Fatal(stages[i])}}
 if nextTeamStage("released")!=""||nextTeamStage("unknown")!=""{t.Fatal("invalid transition")}
 t.Setenv("INKVEIL_TEAM_FILE",filepath.Join(t.TempDir(),"team.json"))
 if e:=saveTeamTasks([]teamTask{{ID:"sample",Stage:"draft"}});e!=nil{t.Fatal(e)}
 got,e:=loadTeamTasks();if e!=nil||len(got)!=1||got[0].Stage!="draft"{t.Fatal(got,e)}
}
