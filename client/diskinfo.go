package client

import (
	"github.com/shirou/gopsutil/v3/disk"
	"strings"
)

//type HDDUsage struct {
//	Size uint64
//	Used uint64
//}

// 该方法在 mac 上存在比较严重的问题，不能获取到容量，增加 apfs 格式后会获取到重复容量
func GetHDDUsage() (HDDSize uint64, HDDUsed uint64) {
	validFs := []string{"ext4", "ext3", "ext2", "reiserfs", "jfs", "btrfs",
		"fuseblk", "zfs", "simfs", "ntfs", "fat32", "exfat", "xfs"}

	partitions, err := disk.Partitions(true)
	//fmt.Printf("par:%v\n", partitions)
	if err != nil {
		return 0, 0
	}
	totalSizeMB := disk.UsageStat{}.Total
	usedSizeMB := disk.UsageStat{}.Total
	for _, part := range partitions {
		if contains(validFs, strings.ToLower(part.Fstype)) {
			usage, err := disk.Usage(part.Mountpoint)
			if err != nil {
				continue
			}
			totalSizeMB += usage.Total / 1024 / 1024
			usedSizeMB += usage.Used / 1024 / 1024
		}
	}
	return totalSizeMB, usedSizeMB
}

// 自定义一个检查切片是否包含某元素的函数
func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
