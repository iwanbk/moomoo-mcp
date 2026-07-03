package moomoo

import (
	"context"
	"fmt"
	"strings"

	"github.com/hyperjiang/futu/adapt"
)

// SecurityGroup is one watchlist group (e.g. a custom group or a system
// group like "Recently Viewed").
type SecurityGroup struct {
	GroupName string `json:"group_name"`
	GroupType string `json:"group_type"`
}

// WatchlistSecurity is one security entry within a watchlist group.
type WatchlistSecurity struct {
	Code      string `json:"code"`
	Name      string `json:"name"`
	LotSize   int32  `json:"lot_size,omitempty"`
	SecType   string `json:"sec_type,omitempty"`
	ListTime  string `json:"list_time,omitempty"`
	Delisting bool   `json:"delisting,omitempty"`
}

var groupTypeNames = map[int32]string{
	adapt.GroupType_Unknown: "UNKNOWN",
	adapt.GroupType_Custom:  "CUSTOM",
	adapt.GroupType_System:  "SYSTEM",
	adapt.GroupType_All:     "ALL",
}

var groupTypeIDs = func() map[string]int32 {
	m := make(map[string]int32, len(groupTypeNames))
	for id, name := range groupTypeNames {
		m[name] = id
	}
	return m
}()

var secTypeNames = map[int32]string{
	adapt.SecurityType_Unknown:  "UNKNOWN",
	adapt.SecurityType_Bond:     "BOND",
	adapt.SecurityType_Bwrt:     "BASKET_WARRANT",
	adapt.SecurityType_Eqty:     "EQUITY",
	adapt.SecurityType_Trust:    "TRUST",
	adapt.SecurityType_Warrant:  "WARRANT",
	adapt.SecurityType_Index:    "INDEX",
	adapt.SecurityType_Plate:    "PLATE",
	adapt.SecurityType_Drvt:     "OPTION",
	adapt.SecurityType_PlateSet: "PLATE_SET",
	adapt.SecurityType_Future:   "FUTURE",
	adapt.SecurityType_Forex:    "FOREX",
	adapt.SecurityType_Crypto:   "CRYPTO",
}

// GetUserSecurityGroup returns the user's watchlist groups. groupType
// selects which kind of group to list (CUSTOM, SYSTEM, ALL); empty defaults
// to ALL.
func (c *Client) GetUserSecurityGroup(ctx context.Context, groupType string) ([]SecurityGroup, error) {
	gt := adapt.GroupType_All
	if groupType != "" {
		id, ok := groupTypeIDs[strings.ToUpper(groupType)]
		if !ok {
			return nil, fmt.Errorf("unknown group_type %q; must be CUSTOM, SYSTEM or ALL", groupType)
		}
		gt = id
	}
	groups, err := c.sdk.GetUserSecurityGroupWithContext(ctx, gt)
	if err != nil {
		return nil, err
	}
	out := make([]SecurityGroup, 0, len(groups))
	for _, g := range groups {
		out = append(out, SecurityGroup{
			GroupName: g.GetGroupName(),
			GroupType: groupTypeNames[g.GetGroupType()],
		})
	}
	return out, nil
}

// GetUserSecurity returns the securities in one watchlist group. groupName
// comes from GetUserSecurityGroup.
func (c *Client) GetUserSecurity(ctx context.Context, groupName string) ([]WatchlistSecurity, error) {
	infos, err := c.sdk.GetUserSecurityWithContext(ctx, groupName)
	if err != nil {
		return nil, err
	}
	out := make([]WatchlistSecurity, 0, len(infos))
	for _, info := range infos {
		b := info.GetBasic()
		if b == nil {
			continue
		}
		out = append(out, WatchlistSecurity{
			Code:      adapt.SecurityToCode(b.GetSecurity()),
			Name:      b.GetName(),
			LotSize:   b.GetLotSize(),
			SecType:   secTypeNames[b.GetSecType()],
			ListTime:  b.GetListTime(),
			Delisting: b.GetDelisting(),
		})
	}
	return out, nil
}
