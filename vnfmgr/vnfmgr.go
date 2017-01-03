package vnfmgr
import (
	"fmt"
	"sync"
	"sort"
	"regexp"
	"strconv"
)

type VnfOp int

type VnfFuture struct {
	name string
	result chan error
}

const (
	VNF_ADMIN_INIT VnfOp = 0 + iota
	VNF_ADMIN_CREATE
	VNF_ADMIN_UPDATE
	VNF_ADMIN_DELETE
	VNF_RANDOM_OP
	VNF_MAX_OPS
)

const (
	VNF_ADMIN_CHANNEL_REQUESTS = 16
)

var (
	vnfOps = [...]string{ "VNF_INIT", "VNF_CREATE", "VNF_UPDATE", "VNF_DELETE", "VNF_RANDOM"}
	vnf_regex *regexp.Regexp = regexp.MustCompile("[0-9]+$")
)

func (op VnfOp) String() string {
	return vnfOps[op]
}

func newVnfFuture(name string) *VnfFuture {
	future := new(VnfFuture)
	future.name = name
	future.result = make(chan error, VNF_ADMIN_CHANNEL_REQUESTS)
	return future
}

//get the result for a VNF operation
func (vnfFuture VnfFuture) Get() error {
	return <- vnfFuture.result
}

func (vnfFuture VnfFuture) Id() string {
	return vnfFuture.name
}

func (vnfFuture *VnfFuture) send(err error) {
	vnfFuture.result <- err
}

//next state transition map from current state
var vnfNextOpTable = map[VnfOp][]VnfOp { VNF_ADMIN_INIT: []VnfOp{VNF_ADMIN_INIT, VNF_ADMIN_CREATE},
	VNF_ADMIN_CREATE: []VnfOp{VNF_ADMIN_UPDATE, VNF_ADMIN_DELETE},
	VNF_ADMIN_UPDATE: []VnfOp{VNF_ADMIN_UPDATE, VNF_ADMIN_DELETE},
	VNF_ADMIN_DELETE: []VnfOp{VNF_ADMIN_INIT, VNF_ADMIN_CREATE},
}

var vnfSmTable = map[VnfOp]func(vnf *Vnf, args interface{}) error {
	VNF_ADMIN_INIT: vnfInit,
	VNF_ADMIN_CREATE:vnfCreate,
	VNF_ADMIN_UPDATE:vnfUpdate,
	VNF_ADMIN_DELETE:vnfDelete,
}

type Vnf struct {
	name string
	attr string
	state VnfOp
	admin_channel chan VnfWork
}

type Vnfs []Vnf

type VnfWork struct {
	state VnfOp
	args interface{}
	future *VnfFuture
}

//sort interface functions for vnfs
func (vnfs Vnfs) Len() int {
	return len(vnfs)
}

func (vnfs Vnfs) Less(i, j int) bool {
	m1 := vnf_regex.FindString(vnfs[i].name)
	m2 := vnf_regex.FindString(vnfs[j].name)
	if m1 != "" && m2 != "" {
		m1_i, _ := strconv.Atoi(m1)
		m2_i, _ := strconv.Atoi(m2)
		return m1_i < m2_i
	}
	return vnfs[i].name < vnfs[j].name
}

func (vnfs Vnfs) Swap(i, j int) {
	vnfs[i], vnfs[j] = vnfs[j], vnfs[i]
}

func (vnf *Vnf) initialize() error {
	vnf.state = VNF_ADMIN_INIT
	go func() {
		for {
			vnf_work, ok := <-vnf.admin_channel
			if !ok { //channel closed
				break
			}
			err := vnf.Sm(vnf_work.state, vnf_work.args)
			vnf_work.future.send(err)
		}
	}()
	return nil
}

func newVnf(name string, attr string) *Vnf {
	vnf := new(Vnf)
	vnf.name = name
	vnf.attr = attr
	vnf.state = VNF_ADMIN_INIT
	vnf.admin_channel = make(chan VnfWork, VNF_ADMIN_CHANNEL_REQUESTS)
	return vnf
}

func (vnf *Vnf) init(args interface{}) error {
	fmt.Println("Inside VNF init for VNF", vnf.name)
	return nil
}

func (vnf *Vnf) create(args interface{}) error {
	fmt.Println("Inside VNF create for VNF", vnf.name)
	return nil
}

func (vnf *Vnf) update(args interface{}) error {
	fmt.Println("Inside VNF update for VNF", vnf.name)
	return nil
}

func (vnf *Vnf) delete(args interface{}) error {
	fmt.Println("Inside VNF delete for VNF", vnf.name)
	close(vnf.admin_channel)
	return nil
}

func (vnf *Vnf) validateTransition(next_state VnfOp) error {
	//validate the vnf op table transition
	next_states, ok := vnfNextOpTable[vnf.state]
	if !ok {
		return fmt.Errorf("Invalid VNF current state: %s", vnf.state)
	}
	for _, state := range next_states {
		if state == next_state {
				return nil
		}
	}
	return fmt.Errorf("VNF %s cannot transition to State %s from State %s",
		vnf.name, next_state, vnf.state)
}

func (vnf *Vnf) Sm(state VnfOp, args interface{}) error {
	cur_state := vnf.state
	if err := vnf.validateTransition(state); err != nil {
		fmt.Println(err)
		return err
	}
	//run the callback
	if err := vnfSmTable[state](vnf, args); err != nil {
		fmt.Println("VNF", vnf.name, "failed to transition to state", state, "from state", cur_state)
		return err
	}
	vnf.state = state
	fmt.Println("VNF", vnf.name, "transitioned from state", cur_state, "to state", state)
	return nil
}

func (vnf *Vnf) worker(op VnfOp, args interface{}) error {
	fmt.Println("Inside random VNF operation", op, "for VNF", vnf.name)
	return nil
}

func vnfInit(vnf *Vnf, args interface{}) error {
	return vnf.init(args)
}

func vnfCreate(vnf *Vnf, args interface{}) error {
	return vnf.create(args)
}

func vnfUpdate(vnf *Vnf, args interface{}) error {
	return vnf.update(args)
}

func vnfDelete(vnf *Vnf, args interface{}) error {
	return vnf.delete(args)
}


type VnfMgr struct {
	numVnfs int
	vnfTable map[string]*Vnf
	mutex sync.Mutex
}

func NewVnfMgr() *VnfMgr {
	vnfmgr := new(VnfMgr)
	vnfmgr.numVnfs = 0
	vnfmgr.vnfTable = make(map[string]*Vnf)
	return vnfmgr
}

func (vnfMgr *VnfMgr) removeVnf(vnf *Vnf) {
	vnfMgr.mutex.Lock()
	delete(vnfMgr.vnfTable, vnf.name)
	vnfMgr.numVnfs--
	vnfMgr.mutex.Unlock()
}

func (vnfMgr *VnfMgr) createVnfAsync(vnf *Vnf, args interface{}) VnfFuture {
	future := newVnfFuture(vnf.name)
	go func() {
		vnfMgr.mutex.Lock()
		_, ok := vnfMgr.vnfTable[vnf.name]
		if ok {
			vnfMgr.mutex.Unlock()
			future.send(fmt.Errorf("VNF %s already exist", vnf.name))
			return
		}
		vnfMgr.vnfTable[vnf.name] = vnf
		vnfMgr.numVnfs++
		vnfMgr.mutex.Unlock()
		if err := vnf.initialize(); err != nil {
			fmt.Println(err)
			vnfMgr.removeVnf(vnf)
			future.send(err)
			return
		}
		vnf_work := VnfWork{state: VNF_ADMIN_CREATE, args: args, future: future}
		vnf.admin_channel <- vnf_work
	} ()
	return *future
}

func (vnfMgr *VnfMgr) createVnf(name string, args interface{}) VnfFuture {
	vnf := newVnf(name, args.(string))
	return vnfMgr.createVnfAsync(vnf, args)
}

func (vnfMgr *VnfMgr) adminVnfAsync(name string, op VnfOp, args interface{}) VnfFuture {
	future := newVnfFuture(name)
	go func() {
		vnfMgr.mutex.Lock()
		vnf, ok := vnfMgr.vnfTable[name]
		if !ok {
			vnfMgr.mutex.Unlock()
			future.send(fmt.Errorf("VNF %s does not exist", name))
			return
		}
		if op == VNF_ADMIN_DELETE {
			delete(vnfMgr.vnfTable, name)
			vnfMgr.numVnfs--
		}
		vnfMgr.mutex.Unlock()
		vnf_work := VnfWork{state: op, args: args, future: future}
		vnf.admin_channel <- vnf_work
	} ()
	return *future
}

func (vnfMgr *VnfMgr) adminVnf(name string, op VnfOp, args interface{}) VnfFuture {
	return vnfMgr.adminVnfAsync(name, op, args)
}

func (vnfMgr *VnfMgr) nonAdminVnfAsync(name string, op VnfOp, args interface{}) VnfFuture {
	future := newVnfFuture(name)
	go func() {
		vnfMgr.mutex.Lock()
		vnf, ok := vnfMgr.vnfTable[name]
		vnfMgr.mutex.Unlock()
		if !ok {
			future.send(fmt.Errorf("VNF %s does not exist", name))
			return
		}
		future.send(vnf.worker(op, args.(string)))
	} ()
	return *future
}

func (vnfMgr *VnfMgr) nonAdminVnf(name string, op VnfOp, args interface{}) VnfFuture {
	return vnfMgr.nonAdminVnfAsync(name, op, args)
}

func (vnfMgr *VnfMgr) Dispatch(name string, op VnfOp, args interface{}) VnfFuture {
	switch {
	case op == VNF_ADMIN_CREATE:
		return vnfMgr.createVnf(name, args)
	case VNF_ADMIN_UPDATE <= op && op <= VNF_ADMIN_DELETE:
		return vnfMgr.adminVnf(name, op, args)
	}
	return vnfMgr.nonAdminVnf(name, op, args)
}

func (vnfMgr *VnfMgr) Create(vnfs []string, args []string) []VnfFuture {
	var futures []VnfFuture
	for i, name := range vnfs {
		futures = append(futures, vnfMgr.Dispatch(name, VNF_ADMIN_CREATE, args[i]))
	}
	return futures
}

func (vnfMgr *VnfMgr) Get(id string) (Vnf, bool) {
	vnfMgr.mutex.Lock()
	defer vnfMgr.mutex.Unlock()
	vnf, ok := vnfMgr.vnfTable[id]
	if ok {
		return *vnf, ok
	}
	return Vnf{}, ok
}

func (vnfMgr *VnfMgr) GetVnfs() Vnfs {
	var vnfs Vnfs
	vnfMgr.mutex.Lock()
	defer vnfMgr.mutex.Unlock()
	for _, vnf := range vnfMgr.vnfTable {
		vnfs = append(vnfs, *vnf)
	}
	sort.Sort(vnfs)
	return vnfs
}
