package main

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"time"

	"crypto/x509"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type Person struct {
	Name  string
	Phone string
}

type Counter struct {
	ID int
}

type MongoController struct {
	URL        string
	SSL        bool
	CaCertFile string
	session    mgo.Session
	database   string
}

// Connect the drive to a MongoDB Instance with the given URL
func (m MongoController) Connect() *mgo.Session {
	dialInfo, err := mgo.ParseURL(m.URL)

	if err != nil {
		panic(err)
	}

	if m.SSL {
		tlsConfig := &tls.Config{}

		if m.CaCertFile != "" {
			roots := x509.NewCertPool()
			if ca, err := ioutil.ReadFile("ca_cert.pem"); err == nil {
				roots.AppendCertsFromPEM(ca)
			}
			tlsConfig.RootCAs = roots
		} else {
			tlsConfig.InsecureSkipVerify = true
		}

		dialInfo.DialServer = func(addr *mgo.ServerAddr) (net.Conn, error) {
			conn, err := tls.Dial("tcp", addr.String(), tlsConfig)
			return conn, err
		}
	}

	session, err := mgo.DialWithInfo(dialInfo)
	if err != nil {
		panic(err)
	}

	// Optional. Switch the session to a monotonic behavior.
	session.SetMode(mgo.Monotonic, true)
	return session
}

func (m MongoController) TestWithSampledata() {
	session := m.Connect()
	defer session.Close()

	c := session.DB("test").C("phone")
	err := c.Insert(&Person{"Ale", "+55 53 8116 9639"},
		&Person{"Cla", "+55 53 8402 8510"})
	if err != nil {
		log.Fatal(err)
	}

	result := Person{}
	err = c.Find(bson.M{"name": "Ale"}).One(&result)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Sample data written successfull...")
}

func (m MongoController) WriteCounter(repetition int) {
	i := 0
	session := m.Connect()
	defer session.Close()
	c := session.DB("test").C("countertest")

	for i < repetition {
		// insert value
		err := c.Insert(&Counter{i})
		if err != nil {
			fmt.Println()
			fmt.Println(err)
			//try one more time
			session = m.Connect()
			c = session.DB("test").C("countertest")
			err = c.Insert(&Counter{i})
			if err != nil {
				panic(err)
			}
		}

		// check if value exists
		counterResult := Counter{}
		err = c.Find(bson.M{"id": i}).One(&counterResult)
		if err != nil {
			log.Fatal(err)
		}
		if i%100 == 0 {
			fmt.Print(".")
		}

		// sleep 100ms
		time.Sleep(100 * time.Millisecond)
		i++
	}
}

func (m MongoController) TestCounter(repetition int) {
	session := m.Connect()
	defer session.Close()
	c := session.DB("newtest").C("countertest")

	// check for all entries
	for i := 0; i < repetition; i++ {
		if i%100 == 0 {
			fmt.Print(".")
		}
		counterResult := Counter{}
		err := c.Find(bson.M{"id": i}).One(&counterResult)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func main() {
	cmd := NewCommandline(os.Args)

	mc := MongoController{
		URL:        cmd.URL,
		SSL:        cmd.SSL,
		CaCertFile: cmd.CaCert,
	}

	mc.TestWithSampledata()

	fmt.Print("Start writting test data")
	mc.WriteCounter(cmd.Repetition)
	fmt.Print(" done!\n")

	fmt.Print("Verify test data")
	mc.TestCounter(cmd.Repetition)
	fmt.Print(" done!\n")
}
