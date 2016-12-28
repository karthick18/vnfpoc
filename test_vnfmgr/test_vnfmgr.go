package main
import (
	"fmt"
	"github.com/karthick18/vnfpoc/vnfmgr"
)

func createVnfs(vnf_mgr *vnfmgr.VnfMgr, vnfs []string, args []string) {
	futures := vnf_mgr.Create(vnfs, args)
	for i, f := range futures {
		if err := f.Get(); err != nil {
			fmt.Println("Error", err, "creating VNF", vnfs[i])
		} else {
			fmt.Println("VNF", vnfs[i], "successfully created")
		}
	}
}

func updateVnfs(vnf_mgr *vnfmgr.VnfMgr, vnfs []string, args []string) {
	var futures []vnfmgr.VnfFuture
	for i, name := range vnfs {
		futures = append(futures, vnf_mgr.Dispatch(name, vnfmgr.VNF_ADMIN_UPDATE, args[i]))
	}
	for i, f := range futures {
		if err := f.Get(); err != nil {
			fmt.Println("Error", err, "while updating VNF", vnfs[i])
		} else {
			fmt.Println("Updated VNF", vnfs[i], "successfully")
		}
	}
}

func deleteVnfs(vnf_mgr *vnfmgr.VnfMgr, vnfs []string, args []string) {
	var futures []vnfmgr.VnfFuture
	for i, name := range vnfs {
		futures = append(futures, vnf_mgr.Dispatch(name, vnfmgr.VNF_ADMIN_DELETE, args[i]))
	}
	for i, f := range futures {
		if err := f.Get(); err != nil {
			fmt.Println("Error", err, "while deleting VNF", vnfs[i])
		} else {
			fmt.Println("Deleted VNF", vnfs[i], "successfully")
		}
	}
}

func main() {
	vnf_mgr := vnfmgr.NewVnfMgr()
	var vnfs []string
	var args []string
	for i := 0 ; i < 10; i++ {
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

