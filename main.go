package main
import (
	"fmt"
	"os"
	//"io/fs"
	"io"
	"log"
	"encoding/csv"
	"path/filepath"
)
//type MeReader io.Reader
type MeReader = io.Reader
type Me interface {
	Speak()
}
type Name string

func (n Name) Speak() {
	fmt.Printf("say something: %s\n",n)
}
func main(){
	for k := range 100 {
		fmt.Printf("Hello , Tang Lin %d \n",k) 
	} 


	var n Me = Name("hehe")
	n.Speak()
    root := "/Users/youyou/Desktop"
	// fsys := os.DirFS(root)
	path := filepath.Join(root, "5-16子游戏表.csv")

	// data,err := fs.ReadFile(fsys,"5-16子游戏表.csv")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	//fmt.Print(string(data))
	file,err := os.Open(path)
	defer file.Close()
	reader := csv.NewReader(file)

	records,err := reader.ReadAll()

	if(err != nil){
		log.Fatal(err)
	}
	for _, row := range records {
		//fmt.Println(row)
		//fmt.Println("\n")
		for _, ele := range row {
			if ele != "" {
				fmt.Println(ele)
			}
			
		}
	}
	
}

