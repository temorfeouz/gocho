package node

import (
	"container/list"
	"fmt"
	"net"
	"sync"
	"time"
)

var (
	nodeMutex sync.Mutex
)

func announceNode(nodeInfo *NodeInfo) {
	address, err := net.ResolveUDPAddr("udp", MULTICAST_ADDRESS)
	if err != nil {
		return
	}

	conn, err := net.DialUDP("udp", nil, address)
	if err != nil {
		return
	}

	for {
		fmt.Println("sending multicast info")

		message, err := NewAnnouncePacket(nodeInfo)
		if err != nil {
			fmt.Println("Could not get announce package")
			fmt.Println(err)
			continue
		}

		conn.Write([]byte(message))
		time.Sleep(ANNOUNCE_INTERVAL_SEC * time.Second)
	}
}

func listenForNodes(nodeList *list.List) {
	address, err := net.ResolveUDPAddr("udp", MULTICAST_ADDRESS)
	if err != nil {
		return
	}

	conn, err := net.ListenMulticastUDP("udp", nil, address)
	if err != nil {
		return
	}

	conn.SetReadBuffer(MULTICAST_BUFFER_SIZE)

	for {
		packet := make([]byte, MULTICAST_BUFFER_SIZE)
		size, udpAddr, err := conn.ReadFromUDP(packet)
		if err != nil {
			fmt.Println(err)
			continue
		}

		nodeInfo, err := ParseAnnouncePacket(size, udpAddr, packet)

		if err != nil {
			fmt.Println(err)
			continue
		}
		fmt.Printf("Received multicast packet from %s Id: %s\n", udpAddr.String(), nodeInfo.Id)

		go announcedNodeHandler(nodeInfo, nodeList)
	}
}

func announcedNodeHandler(nodeInfo *NodeInfo, nodeList *list.List) {
	nodeMutex.Lock()
	updateNodeList(nodeInfo, nodeList)
	nodeMutex.Unlock()

	fmt.Println("Printing nodes")

	fmt.Print("[")
	for el := nodeList.Front(); el != nil; el = el.Next() {
		fmt.Print(el.Value.(*NodeInfo).Id, " ")
	}
	fmt.Print("]\n\n")
}

func updateNodeList(nodeInfo *NodeInfo, nodeList *list.List) {
	nodeExists := false
	for el := nodeList.Front(); el != nil; el = el.Next() {
		tmp := el.Value.(*NodeInfo)

		// Already in list
		if tmp.Id == nodeInfo.Id {
			tmp.LastMulticast = time.Now().Unix()
			fmt.Printf("Updating node %s multicast\n", nodeInfo.Id)
			nodeExists = true
			break
		}

	}

	for el := nodeList.Front(); el != nil; el = el.Next() {
		tmp := el.Value.(*NodeInfo)
		if isNodeExpired(tmp, EXPIRE_TIMEOUT_SEC) {
			fmt.Println("Node expired, removing: ", tmp.Id)
			nodeList.Remove(el)
		}
	}

	if !nodeExists {
		fmt.Printf("Adding new node! %p %s\n", nodeInfo, nodeInfo.Id)
		nodeInfo.LastMulticast = time.Now().Unix()
		nodeList.PushBack(nodeInfo)
	}
}

func isNodeExpired(nodeInfo *NodeInfo, timeout int) bool {
	diff := time.Now().Unix() - nodeInfo.LastMulticast
	return diff > int64(timeout)
}
