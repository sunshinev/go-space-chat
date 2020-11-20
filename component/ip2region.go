package component

import (
	"log"
	"strings"

	"github.com/lionsoul2014/ip2region/binding/golang/ip2region"
)

type IpSearch struct {
	Region *ip2region.Ip2Region
}

// 初始化ip2region
func InitIpSearch() *IpSearch {
	region, err := ip2region.New("config/ip2region.db")
	if err != nil {
		log.Fatalf("ip search init err %v", err)
	}

	return &IpSearch{
		Region: region,
	}
}

// 转换ip
func (s *IpSearch) Search(ip string) (*ip2region.IpInfo, error) {
	ip = ip[0:strings.LastIndex(ip, ":")]
	ipInfo, err := s.Region.BtreeSearch(ip)
	if err != nil {
		log.Printf("btree search err %v %v", ip, err)
		return nil, err
	}

	return &ipInfo, nil
}
