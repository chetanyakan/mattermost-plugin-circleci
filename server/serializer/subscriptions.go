package serializer

import "github.com/thoas/go-funk"

type StringSubscription map[string]Subscription

type Subscriptions struct {
	ByChannelID map[string]StringSubscription // store the list of subscriptions for a channelID
	ByKey       map[string][]string           // stores the list of channelIDs to which the message needs to be posted for a subscription
}

func NewSubscriptions() *Subscriptions {
	return &Subscriptions{
		ByChannelID: map[string]StringSubscription{},
		ByKey:       map[string][]string{},
	}
}

// Add adds a new subscription to the list of all subscriptions
func (list *Subscriptions) Add(s Subscription) {
	key := s.GetKey()
	if _, contains := list.ByKey[key]; !contains {
		list.ByKey[key] = make([]string, 0)
	}

	if !funk.Contains(list.ByKey[key], s.ChannelID) {
		list.ByKey[key] = append(list.ByKey[key], s.ChannelID)
	}

	if _, found := list.ByChannelID[s.ChannelID]; !found {
		list.ByChannelID[s.ChannelID] = make(StringSubscription)
	}

	list.ByChannelID[s.ChannelID][key] = s
}

// Remove removes a subscription from the list of all subscriptions
func (list *Subscriptions) Remove(s Subscription) {
	key := s.GetKey()
	delete(list.ByChannelID[s.ChannelID], key)
	list.ByKey[key] = funk.FilterString(list.ByKey[key], func(el string) bool {
		return el == s.ChannelID
	})
}

// GetChannelID returns the channelID to which the message for a subscription should be posted to
func (list *Subscriptions) GetChannelIDs(s Subscription) []string {
	return list.ByKey[s.GetKey()]
}

// List returns the list for a particular channel as a formatted mattermost message
func (list *Subscriptions) List(channelID string) []Subscription {
	values := make([]Subscription, 0, len(list.ByChannelID[channelID]))
	for _, v := range list.ByChannelID[channelID] {
		values = append(values, v)
	}

	return values
}
