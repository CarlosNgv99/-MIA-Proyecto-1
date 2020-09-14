%{
package parser

import (
	"regexp"
  "fmt"
  "bytes"
  "io"
  "bufio"
	"os"
	"MIA-P1/actions"
	"strings"
)

var newDisk actions.Disk = actions.Disk{}
var newPartition actions.Partition = actions.Partition{}
var newFDisk actions.FDISK = actions.FDISK{}
var newMount actions.Mount = actions.Mount{}
var newUnmount actions.Unmount = actions.Unmount{}
var newRep actions.Rep = actions.Rep{}
var stringAux string = ""
var skipFound bool = false


type node struct {
  name string
  children []node
}

func (n node) String() string {
  buf := new(bytes.Buffer)
  n.print(buf, " ")
  return buf.String()
}

func (n node) print(out io.Writer, indent string) {
  fmt.Fprintf(out, "\n%v%v", indent, n.name)
  for _, nn := range n.children { nn.print(out, indent + "  ") }
}

func Node(name string) node { return node{name: name} }
func (n node) append(nn...node) node { n.children = append(n.children, nn...); return n }


%}

%union{
    node node
    token string
}

// Terminals 

%token <token> arrow
%token <token> add
%token <token> delete
%token <token> digit
%token <token> digits
%token <token> diskName
%token <token> equals
%token <token> exec
%token <token> fit
%token <token> fdisk
%token <token> greater
%token <token> hyphen
%token <token> idn 
%token <token> id
%token <token> less
%token <token> mia_file
%token <token> mount
%token <token> mount_name
%token <token> mbr
%token <token> mkfs
%token <token> mkdisk
%token <token> number
%token <token> negNumber
%token <token> name
%token <token> path
%token <token> pause
%token <token> rmdisk
%token <token> size
%token <token> tpe
%token <token> unit
%token <token> unmount
%token <token> rep
%token <token> read
%token <token> route
%token <token> ruta
%token <token> quote


%type <token> add
%type <token> arrow
%type <token> delete
%type <token> digit
%type <token> digits
%type <token> diskName
%type <token> equals
%type <token> exec
%type <token> fit
%type <token> fdisk
%type <token> greater
%type <token> hyphen
%type <token> idn 
%type <token> id
%type <token> less
%type <token> mia_file
%type <token> mount
%type <token> mount_name
%type <token> mbr
%type <token> mkfs
%type <token> mkdisk
%type <token> number
%type <token> negNumber
%type <token> name
%type <token> path
%type <token> pause
%type <token> rmdisk
%type <token> size
%type <token> tpe
%type <token> unit
%type <token> unmount
%type <token> rep
%type <token> read
%type <token> route
%type <token> ruta
%type <token> quote



// Non terminals
%type <node> MOUNT
%type <node> MOUNTO
%type <node> MOUNTT
%type <node> PAUSE
%type <node> UNMOUNT
%type <node> EXEC
%type <node> MKDISK
%type <node> MKDISKO
%type <node> MKDISKT
%type <node> RMDISK
%type <node> FDISK
%type <node> FDISKO
%type <node> FDISKT
%type <node> MKFS
%type <node> INSTRUCTION
%type <node> INSTRUCTIONS
%type <node> READ
%type <node> REP 
%type <node> REPO
%type <node> REPT


%%

start: INSTRUCTIONS;

INSTRUCTIONS: INSTRUCTION
			| INSTRUCTIONS INSTRUCTION;

INSTRUCTION: MOUNT
			| UNMOUNT
			| EXEC
			| MKDISK
			| RMDISK
			| FDISK
			| MKFS
			| PAUSE
			| READ
			| REP
			;

PAUSE: pause { actions.PauseAction() };

READ: read arrow route { actions.ReadFile($3)}

EXEC: exec hyphen path arrow route {Exec($5)};

MKDISK: mkdisk MKDISKO { 
	newDisk.CreateDisk()
	newDisk = actions.Disk{}
 }

MKDISKO: MKDISKT 
| MKDISKO MKDISKT EMPTY;

MKDISKT: hyphen size arrow digit { newDisk.SetDiskSize($4) }
| hyphen path arrow route { newDisk.SetDiskRoute($4) }
| hyphen name arrow diskName { newDisk.SetDiskName($4) }
| hyphen path arrow quote route quote { newDisk.SetDiskRoute($5) }
| hyphen unit arrow id { newDisk.SetDiskUnit($4) }
;


REP: rep REPO { 
	newRep.CreateRep()
	newRep = actions.Rep{}
}

REPO: REPT
	| REPO REPT EMPTY;

REPT:hyphen id arrow id  { newRep.SetRepID($4) }
	| hyphen name arrow id { newRep.SetRepName($4) }
	| hyphen path arrow quote route quote { newRep.SetRepPath($5) }
	| hyphen path arrow route { newRep.SetRepPath($4) }
	| hyphen ruta arrow quote route quote { newRep.SetRepRoute($5) }
	| hyphen ruta arrow route { newRep.SetRepRoute($4) }
	;



MOUNT: mount MOUNTO { 
	newMount.SetMount()
	newMount = actions.Mount{}
 }
 | mount { actions.ShowMountedPartitions()};

MOUNTO: MOUNTT
	|	MOUNTO MOUNTT EMPTY;

MOUNTT: hyphen path arrow quote route quote { newMount.SetMountRoute($5) }
	|	hyphen name arrow id { newMount.SetMountName($4) };

UNMOUNT: unmount UNMOUNTO  { 
	newUnmount.UnmountPartition()
	newUnmount = actions.Unmount{}
 };

 UNMOUNTO: UNMOUNTT
	| UNMOUNTO UNMOUNTT EMPTY;

UNMOUNTT: hyphen id arrow id { newUnmount.SetUnmount($2, $4) } ;

FDISK: fdisk FDISKO {
	newFDisk.CreatePartition()
	newPartition = actions.Partition{}
	newFDisk = actions.FDISK{}
}
;

FDISKO: FDISKT 
| FDISKO FDISKT EMPTY;

FDISKT:	hyphen unit arrow id  { newFDisk.SetFUnit($4) }
| hyphen tpe arrow id { newFDisk.SetPartitionType($4) }
| hyphen fit arrow id { newFDisk.SetPartitionFit($4) }
| hyphen delete arrow id { newFDisk.SetDeleteOption($4) }
| hyphen add arrow digit { newFDisk.SetAddOption($4) }
| hyphen size arrow digit  { newFDisk.SetPSize($4) }
| hyphen name arrow id { newFDisk.SetPartitionName($4) }
| hyphen path arrow quote route quote { newFDisk.SetPartitionRoute($5) }
| hyphen path arrow route { newFDisk.SetPartitionRoute($4) }

;


RMDISK: rmdisk hyphen path arrow quote route quote {actions.RemoveDisk($6)};



MKFS: mkfs hyphen id arrow idn {$$ = Node($1)}
	| mkfs hyphen id arrow idn hyphen tpe arrow id {$$ = Node($1)}
;

EMPTY: ;


%% 

// Run exported
func Run() {
	fi := bufio.NewReader(os.NewFile(0, "stdin"))
	yyDebug = 0
	yyErrorVerbose = true
	for {
		var eqn string
		var ok bool

		fmt.Printf(">> ")
		if eqn, ok = input(fi); ok {
			l := newLexer(bytes.NewBufferString(eqn), os.Stdout, "file.name")
			yyParse(l)
		} else {
			break
		}
	}

}

// RunExec exported
func RunExec(reader *bufio.Reader) {
	yyDebug = 0
	yyErrorVerbose = true
	var stringAux string = ""

	for {
		var ok bool
		var eqn string
		
		if eqn, ok = input(reader); ok {
			eqn = strings.TrimSpace(eqn)
			eqn = strings.Replace(eqn," ","",-1)
			if strings.HasPrefix(eqn, "#"){
				continue
			}
			if len(eqn) == 0 || len(eqn) == 1 {
				continue
			}
			if skipFound == true {
				stringAux = stringAux+ eqn 
				stringAux = strings.Replace(stringAux," ","",-1)
				fmt.Println(">>", stringAux)
				l := newLexer(bytes.NewBufferString(stringAux), os.Stdout, "file.name")
				stringAux = ""
				skipFound = false
				yyParse(l)
				continue
			}
			if strings.HasSuffix(eqn,"\\*") {
				stringAux = eqn
				stringAux = strings.TrimRight(stringAux,"\\*")
				skipFound = true
				continue
			}
			fmt.Println(">>", eqn)
			l := newLexer(bytes.NewBufferString(eqn), os.Stdout, "file.name")
			yyParse(l)
			
		} else {
			break
		}
	}
}

func Exec(route string) {
	re, _ := regexp.Compile(`[a-zA-Z]([a-zA-Z]|[0-9])*\.mia`)
	diskName := re.FindString(route)
	if len(diskName) == 0 {
		fmt.Println(">> No file found. Try again.")
		return
	}
	file, err := os.Open(route)
	if err != nil {
		fmt.Println(">> Couldn't read file. Try again.")
		return
	}
	defer file.Close()
	reader := bufio.NewReader(file)
	RunExec(reader)
}


func input(fi *bufio.Reader) (string, bool) {
	s, err := fi.ReadString('\n')
	if err != nil {
		return "", false
	}
	return s, true
}