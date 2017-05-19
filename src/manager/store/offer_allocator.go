package store

import "github.com/Sirupsen/logrus"

func (zk *ZKStore) CreateOfferAllocatorItem(item *OfferAllocatorItem) error {
	bs, err := encode(item)
	if err != nil {
		return err
	}

	path := keyAllocator + "/" + item.SlotID
	return zk.createAll(path, bs)
}

func (zk *ZKStore) DeleteOfferAllocatorItem(slotID string) error {
	return zk.del(keyAllocator + "/" + slotID)
}

func (zk *ZKStore) ListOfferallocatorItems() []*OfferAllocatorItem {
	ret := make([]*OfferAllocatorItem, 0, 0)

	nodes, err := zk.list(keyAllocator)
	if err != nil {
		logrus.Errorln("zk ListOfferallocatorItems error:", err)
		return ret
	}

	for _, node := range nodes {
		if allo := zk.GetOfferallocatorItem(node); allo != nil {
			ret = append(ret, allo)
		}
	}

	return ret
}

func (zk *ZKStore) GetOfferallocatorItem(slotID string) *OfferAllocatorItem {
	bs, err := zk.get(keyAllocator + "/" + slotID)
	if err != nil {
		logrus.Errorln("zk GetOfferallocatorItem error:", err)
		return nil
	}

	alloc := new(OfferAllocatorItem)
	if err := decode(bs, &alloc); err != nil {
		logrus.Errorln("zk GetOfferallocatorItem.decode error:", err)
		return nil
	}

	return alloc
}
