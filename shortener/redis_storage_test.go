package shortener

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/teris-io/shortid"
)

var localRedis = ""

const testAppid = "supwisdom-test"

const testURL = "http://sh.sina.com.cn/news/k/2020-09-14/detail-iivhvpwy6563308.shtml"

func TestMain(m *testing.M) {
	mr, err := miniredis.Run()

	if err != nil {
		log.Fatalf("run mini redis error %v", err)
	}
	log.Println("redis ", mr.Addr())
	localRedis = fmt.Sprintf("redis://%s", mr.Addr())

	os.Exit(m.Run())
}
func TestConnect(t *testing.T) {
	r := NewRedisStorage()
	url, err := url.Parse(localRedis)
	if err != nil {
		t.Fatalf("URL error %v", err)
	}
	err = r.Open(url)
	if err != nil {
		t.Errorf("Connect error %v", err)
	}

	r.Close()
}

func initialzeRedis(t *testing.T) *RedisStorage {
	r := NewRedisStorage()
	url, err := url.Parse(localRedis)
	if err != nil {
		t.Fatalf("URL error %v", err)
	}
	err = r.Open(url)
	if err != nil {
		t.Fatalf("Connect error %v", err)
	}
	return r
}

func remoteClient(t *testing.T, r *RedisStorage) {
	err := r.UnregisterAppID(testAppid)
	if err != nil {
		t.Fatalf("remove appid error %v", err)
	}
}
func TestCreateClient(t *testing.T) {
	r := initialzeRedis(t)
	defer r.Close()
	key, err := r.RegisterAppID(testAppid)
	if err != nil {
		t.Fatalf("create appid error, %v", err)
		return
	}

	defer remoteClient(t, r)
	if len(key) == 0 {
		t.Fatal("key is empty")
		return
	}

	key, err = r.RegisterAppID(testAppid)
	if err == nil {
		t.Fatalf("re-register appid must be failed!")
		return
	}
	t.Logf("re-register appid error %v", err)
}

func TestURL(t *testing.T) {
	r := initialzeRedis(t)
	defer r.Close()

	_, err := r.RegisterAppID(testAppid)
	if err != nil {
		t.Fatalf("create appid error, %v", err)
		return
	}

	defer remoteClient(t, r)

	var id string
	id, _ = shortid.Generate()
	err = r.NewURL(testURL, id, testAppid, 180)
	if err != nil {
		t.Fatalf("create url error %v", err)
		return
	}
	defer r.DeleteURLByID(id)

	err = r.NewURL(testURL, id, testAppid, 180)
	if err == nil {
		t.Fatalf("re-create url must be error")
		return
	}

	var find string
	find, err = r.FindByID(id)
	if err != nil {
		t.Fatalf("can't find url, error %v", err)
		return
	}
	if find != testURL {
		t.Fatalf("url not matched")
		return
	}

	var newid string
	newid, _ = shortid.Generate()
	err = r.NewURL(testURL, newid, testAppid, 180)

	if err != nil {
		t.Fatalf("re-create url with new id  error %v", err)
		return
	}
}
