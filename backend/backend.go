package backend

import (
	"fmt"
	"github.com/go-redis/redis"
	"time"
)

const (
	kUsers             = "users"
	kUserInfoPrefix    = "user"
	kUserInfoKeyName   = "name"
	kUserInfoKeyPasswd = "passwd"
)

func userHash(name string) string {
	return fmt.Sprintf("%s:%s", kUserInfoPrefix, name)
}

type UserInfo struct {
	Name,
	Password string
}

type UserList []string

func getUserList(client *redis.Client) (UserList, error) {
	numUsers, err := client.LLen(kUsers).Result()
	if err != nil {
		return UserList{}, err
	}

	userNames, err := client.LRange(kUsers, 0, numUsers).Result()
	if err != nil {
		return UserList{}, err
	}
	return UserList(userNames), nil
}

func getUserInfo(client *redis.Client, username string) (UserInfo, error) {
	info, err := client.HMGet(userHash(username), kUserInfoKeyName, kUserInfoKeyPasswd).Result()
	if err != nil {
		return UserInfo{}, err
	}
	return UserInfo{
		Name:     info[0].(string),
		Password: info[1].(string),
	}, nil

}

type Client struct {
	redisAddress string
	dialTimeout  time.Duration
}

func NewClient(address string, timeout time.Duration) *Client {
	return &Client{
		redisAddress: address,
		dialTimeout:  timeout,
	}
}

func (c *Client) SeedData(data map[string]UserInfo) error {
	client := redis.NewClient(&redis.Options{Addr: c.redisAddress})
	defer client.Close()

	if err := client.Del(kUsers).Err(); err != nil {
		return fmt.Errorf("failed to delete existing user list: %v", err)
	}

	for name, info := range data {
		err := client.LPush(kUsers, name).Err()
		if err != nil {
			return fmt.Errorf("failed to insert user %q: %v", name, err)
		}
		if err := client.HMSet(userHash(name), map[string]interface{}{
			kUserInfoKeyName:   info.Name,
			kUserInfoKeyPasswd: info.Password,
		}).Err(); err != nil {
			return fmt.Errorf("failed to insert user info for %q: %v", name, err)
		}
	}
	return nil
}

func (c *Client) GetUserNames() (UserList, error) {
	client := redis.NewClient(&redis.Options{Addr: c.redisAddress, DialTimeout: c.dialTimeout})
	defer client.Close()

	return getUserList(client)
}

func (c *Client) GetAllUserInfo() (map[string]UserInfo, error) {
	client := redis.NewClient(&redis.Options{Addr: c.redisAddress, DialTimeout: c.dialTimeout})
	defer client.Close()

	userNames, err := getUserList(client)
	if err != nil {
		return map[string]UserInfo{}, fmt.Errorf("failed to get user list: %v", err)
	}

	userInfoMap := make(map[string]UserInfo, len(userNames))
	for _, name := range userNames {
		info, err := getUserInfo(client, name)
		if err != nil {
			return map[string]UserInfo{}, fmt.Errorf("failed to get user info for %q: %v", name, err)
		}
		userInfoMap[name] = info
	}
	return userInfoMap, nil
}
