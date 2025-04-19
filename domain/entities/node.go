package entities

type Node struct {
	Name            string
	IP              string
	Cores           int64
	Memory          int64
	MemoryAllocated int64
	Disk            int64
	DiskAllocated   int64
	Role            string
	TaskCount       int
}

func NewNode(name, api, role string) *Node {
	return &Node{
		Name:            name,
		IP:              api,
		Cores:           0,
		Memory:          0,
		MemoryAllocated: 0,
		Disk:            0,
		DiskAllocated:   0,
		Role:            role,
		TaskCount:       0,
	}
}
