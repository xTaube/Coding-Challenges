// https://codingchallenges.fyi/challenges/challenge-load-balancer/#the-challenge---building-a-load-balancer

package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"time"

	"github.com/xTaube/coding-challenges/load-balancer/src/http"
)

const BACKEND_SERVICE_ADDRESSES = "127.0.0.1:8081,127.0.0.1:8082,127.0.0.1:8083"


type InvalidServiceAddressError struct {
	address string
}

func(err *InvalidServiceAddressError) Error() string {
	return fmt.Sprintf("%s is not valid service address.", err.address)
}

type RequestForwardError struct {
	originErr error
}

func(err *RequestForwardError) Error() string {
	return err.originErr.Error()
}

type NoAvailableServiceInstance struct {}

func(err *NoAvailableServiceInstance) Error() string {
	return "All service instances are unavailable."
}

type ServiceInstance struct {
	host string
	port string
	isHealthy bool
}

func (instance *ServiceInstance) Addr() string {
	return fmt.Sprintf("%s:%s", instance.host, instance.port)
}

func (instance *ServiceInstance) openConnection() (net.Conn, error) {
	connection, err := net.Dial("tcp", instance.Addr())
	if err != nil {
		return nil, err
	}
	return connection, nil
}

func (instance *ServiceInstance) healthCheck() {
	connection, err := instance.openConnection()
	if err != nil {
		log.Println(err)
		instance.isHealthy = false
		return
	}

	_, err = connection.Write([]byte(fmt.Sprintf("GET /health HTTP/1.1\r\nHost: %s\r\nAccept: */*\r\n", instance.Addr())))
	if err != nil {
		log.Println(err)
		instance.isHealthy = false
		return
	}
	
	response, err := http.ReadResponse(connection)
	if err != nil {
		log.Println(err)
		instance.isHealthy = false
		return
	}
	instance.isHealthy = response.Status() == 200
}

func (instance *ServiceInstance) Forward(request []byte, response []byte) error {
	connection, err := instance.openConnection()
	
	if err != nil {
		return err
	}
	
	defer connection.Close()

	_, err = connection.Write(request)

	if err != nil {
		log.Printf("Could not forward request to backend service %s\n", instance.Addr())
		return &RequestForwardError{err}
	}

	_, err = connection.Read(response)

	if err != nil {
		log.Printf("Could not recieve response from backend service %s\n", instance.Addr())

		return &RequestForwardError{err}
	}
	return nil
}

type LoadBalancer struct {
	instances []*ServiceInstance
	instancesNumber int
	nextRequestForwardInstance int
}

func initLoadBalancer() *LoadBalancer {
	return &LoadBalancer{
		instances: []*ServiceInstance{},
		instancesNumber: 0,
		nextRequestForwardInstance: 0,
	}
}

func(lb *LoadBalancer) registerInstance(address string) error {
	hostPortPair := strings.Split(address, ":")

	if len(hostPortPair) != 2 {
		return &InvalidServiceAddressError{address}
	}

	instance := ServiceInstance{hostPortPair[0], hostPortPair[1], true}

	lb.instances = append(lb.instances, &instance)
	lb.instancesNumber++

	log.Printf("Service instance %s successfully registered in load balancer\n", instance.Addr())

	return nil
}

func(lb *LoadBalancer) servicesHealthCheck(period time.Duration) {
	for {
		for _, instance := range lb.instances {
			instance.healthCheck()
			log.Printf("%s availability: %t", instance.Addr(), instance.isHealthy)
		}
		time.Sleep(period*time.Second)
	}
}

func(lb *LoadBalancer) getNextServiceInstance() *ServiceInstance {
	instance := lb.instances[lb.nextRequestForwardInstance]

	lb.nextRequestForwardInstance = (lb.nextRequestForwardInstance + 1) % lb.instancesNumber

	return instance
}

func(lb *LoadBalancer) getAvailableServiceInstance() (*ServiceInstance, error) {
	var instance *ServiceInstance
	
	for i:=0; i<lb.instancesNumber; i++ {
		instance = lb.getNextServiceInstance()
		if instance.isHealthy {
			return instance, nil
		}
	}
	
	return nil, &NoAvailableServiceInstance{}
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("You need to specify port.\n")
	}
	port := os.Args[1]

	server, err := net.Listen("tcp", fmt.Sprintf(":%s", port))

	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Load balancer listening on %s\n", server.Addr())

	loadBalancer := initLoadBalancer()
	for _, address := range strings.Split(BACKEND_SERVICE_ADDRESSES, ",") {
		if err := loadBalancer.registerInstance(address); err != nil {
			log.Fatal(err)
		}
	}

	go loadBalancer.servicesHealthCheck(10)

	for {
		client, err := server.Accept()

		if err != nil {
			log.Println("Could not connect with client.")
			continue
		}

		go handleClient(client, loadBalancer)
	}
}


func handleClient(client net.Conn, loadBalancer *LoadBalancer) {
	defer client.Close()

	requestBuff := make([]byte, 1024)
	n, err := client.Read(requestBuff)

	log.Printf("Request bytes: %d\n", n)

	if err != nil {
		log.Printf("Could not read request from client %s\n", client.RemoteAddr())
		return
	}

	log.Printf("Recived request from %s\n%s\n", client.RemoteAddr(), requestBuff)

	instance, err := loadBalancer.getAvailableServiceInstance()

	if err != nil {
		log.Println(err)
		client.Write([]byte("HTTP/1.1 500 INTERNAL SERVER ERROR\r\n"))
		return
	}

	responseBuff := make([]byte, 1024)
	err = instance.Forward(requestBuff, responseBuff)

	if err != nil {
		client.Write([]byte("HTTP/1.1 500 INTERNAL SERVER ERROR\r\n"))
		return
	}

	client.Write(responseBuff)
}