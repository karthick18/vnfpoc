package main
import (
	"fmt"
	"github.com/karthick18/vnfpoc/vnfmgr"
	"os"
	"strconv"
)

func createVnfs(vnf_mgr *vnfmgr.VnfMgr, vnfs []string, args []string) {
	futures := vnf_mgr.Create(vnfs, args)
	for _, f := range futures {
		if err := f.Get(); err != nil {
			fmt.Println("Error", err, "creating VNF", f.Id())
		} else {
			fmt.Println("VNF", f.Id(), "successfully created")
		}
	}
}

func updateVnfs(vnf_mgr *vnfmgr.VnfMgr, vnfs []string, args []string) {
	var futures []vnfmgr.VnfFuture
	for i, name := range vnfs {
		futures = append(futures, vnf_mgr.Dispatch(name, vnfmgr.VNF_ADMIN_UPDATE, args[i]))
	}
	for _, f := range futures {
		if err := f.Get(); err != nil {
			fmt.Println("Error", err, "while updating VNF", f.Id())
		} else {
			fmt.Println("Updated VNF", f.Id(), "successfully")
		}
	}
}

func deleteVnfs(vnf_mgr *vnfmgr.VnfMgr, vnfs []string, args []string) {
	var futures []vnfmgr.VnfFuture
	for i, name := range vnfs {
		futures = append(futures, vnf_mgr.Dispatch(name, vnfmgr.VNF_ADMIN_DELETE, args[i]))
	}
	for _, f := range futures {
		if err := f.Get(); err != nil {
			fmt.Println("Error", err, "while deleting VNF", f.Id())
		} else {
			fmt.Println("Deleted VNF", f.Id(), "successfully")
		}
	}
}

func main() {
	vnf_mgr := vnfmgr.NewVnfMgr()
	var vnfs []string
	var args []string
	numVnfs := 10
	if len(os.Args) > 1 {
		numVnfs, _ = strconv.Atoi(os.Args[1])
		if numVnfs <= 0 {
			numVnfs = 10
		}
	}
	fmt.Println("Testing with", numVnfs, "VNFS")
	for i := 0 ; i < numVnfs; i++ {
		name := fmt.Sprintf("vnf_%d", i)
		vnfs = append(vnfs, name)
		args = append(args, name)
	}
	createVnfs(vnf_mgr, vnfs, args)
	updateVnfs(vnf_mgr, vnfs, args)
	deleteVnfs(vnf_mgr, vnfs, args)
	//force an invalid transition
	updateVnfs(vnf_mgr, vnfs, args)
}

