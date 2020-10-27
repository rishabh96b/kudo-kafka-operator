package service

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"testing"

	"github.com/onsi/ginkgo/reporters"

	"github.com/golang/mock/gomock"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/mesosphere/kudo-kafka-operator/images/kafka-utils/pkgs/mocks"

	. "github.com/onsi/gomega"

	. "github.com/onsi/ginkgo"
	testclient "k8s.io/client-go/kubernetes/fake"
)

var _ = Describe("[Kafka KafkaService]", func() {

	var (
		mockCtrl *gomock.Controller
		mockEnv  *mocks.MockEnvironment
	)

	tests := []struct {
		svc                                 *v1.ServiceList
		node                                *v1.Node
		name                                string
		expectedAdvertisedListeners         string
		expectedListeners                   string
		expectedListenerSecurityProtocolMap string
		expectedExternalDNS                 string
		nodeportIpType                      string
	}{
		{
			name: "Type LoadBalancer AWS",
			svc: &v1.ServiceList{
				Items: []v1.Service{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "kafka-kafka-0-external",
							Namespace: v1.NamespaceDefault,
						},
						Spec: v1.ServiceSpec{
							Type: v1.ServiceTypeLoadBalancer,
						},
						Status: v1.ServiceStatus{
							LoadBalancer: v1.LoadBalancerStatus{
								Ingress: []v1.LoadBalancerIngress{
									{
										Hostname: "aws.kafka.dns-kafka-kafka-0",
									},
								},
							},
						},
					},
				},
			},
			node:                                &v1.Node{},
			expectedAdvertisedListeners:         "EXTERNAL_INGRESS://aws.kafka.dns-kafka-kafka-0:9097",
			expectedListeners:                   "EXTERNAL_INGRESS://0.0.0.0:9097",
			expectedExternalDNS:                 "aws.kafka.dns-kafka-kafka-0",
			expectedListenerSecurityProtocolMap: "EXTERNAL_INGRESS:PLAINTEXT",
		},
		{
			name: "Type LoadBalancer GCE",
			svc: &v1.ServiceList{
				Items: []v1.Service{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "kafka-kafka-0-external",
							Namespace: v1.NamespaceDefault,
						},
						Spec: v1.ServiceSpec{
							Type: v1.ServiceTypeLoadBalancer,
						},
						Status: v1.ServiceStatus{
							LoadBalancer: v1.LoadBalancerStatus{
								Ingress: []v1.LoadBalancerIngress{
									{
										IP: "30.0.0.1",
									},
								},
							},
						},
					},
				},
			},
			node:                                &v1.Node{},
			expectedAdvertisedListeners:         "EXTERNAL_INGRESS://30.0.0.1:9097",
			expectedListeners:                   "EXTERNAL_INGRESS://0.0.0.0:9097",
			expectedExternalDNS:                 "30.0.0.1",
			expectedListenerSecurityProtocolMap: "EXTERNAL_INGRESS:PLAINTEXT",
		},
		{
			name: "Type NodePort External IP",
			svc: &v1.ServiceList{
				Items: []v1.Service{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "kafka-kafka-0-external",
							Namespace: v1.NamespaceDefault,
						},
						Spec: v1.ServiceSpec{
							Type: v1.ServiceTypeNodePort,
							Ports: []v1.ServicePort{
								{
									Port:     31002,
									NodePort: 31002,
								},
							},
						},
						Status: v1.ServiceStatus{
							LoadBalancer: v1.LoadBalancerStatus{
								Ingress: []v1.LoadBalancerIngress{
									{
										IP: "10.0.0.1",
									},
								},
							},
						},
					},
				},
			},
			node: &v1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "kubelet-0",
				},
				Status: v1.NodeStatus{
					Addresses: []v1.NodeAddress{
						{
							Type:    v1.NodeExternalIP,
							Address: "30.0.0.1",
						},
					},
				},
			},
			expectedAdvertisedListeners:         "EXTERNAL_INGRESS://30.0.0.1:31002",
			expectedListeners:                   "EXTERNAL_INGRESS://0.0.0.0:31002",
			expectedExternalDNS:                 "30.0.0.1",
			expectedListenerSecurityProtocolMap: "EXTERNAL_INGRESS:PLAINTEXT",
			nodeportIpType:                      "EXTERNAL",
		},
		{
			name: "Type NodePort Internal IP",
			svc: &v1.ServiceList{
				Items: []v1.Service{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "kafka-kafka-0-external",
							Namespace: v1.NamespaceDefault,
						},
						Spec: v1.ServiceSpec{
							Type: v1.ServiceTypeNodePort,
							Ports: []v1.ServicePort{
								{
									Port:     31002,
									NodePort: 31002,
								},
							},
						},
						Status: v1.ServiceStatus{
							LoadBalancer: v1.LoadBalancerStatus{
								Ingress: []v1.LoadBalancerIngress{
									{
										IP: "10.0.0.1",
									},
								},
							},
						},
					},
				},
			},
			node: &v1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "kubelet-0",
				},
				Status: v1.NodeStatus{
					Addresses: []v1.NodeAddress{
						{
							Type:    v1.NodeInternalIP,
							Address: "10.0.0.1",
						},
					},
				},
			},
			expectedAdvertisedListeners:         "EXTERNAL_INGRESS://10.0.0.1:31002",
			expectedListeners:                   "EXTERNAL_INGRESS://0.0.0.0:31002",
			expectedExternalDNS:                 "10.0.0.1",
			expectedListenerSecurityProtocolMap: "EXTERNAL_INGRESS:PLAINTEXT",
			nodeportIpType:                      "INTERNAL",
		},
		{
			name: "No external Service",
			svc:  &v1.ServiceList{},
			node: &v1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "kubelet-0",
				},
				Status: v1.NodeStatus{
					Addresses: []v1.NodeAddress{
						{
							Type:    v1.NodeExternalIP,
							Address: "30.0.0.1",
						},
					},
				},
			},
			expectedAdvertisedListeners:         "",
			expectedListeners:                   "",
			expectedExternalDNS:                 "",
			expectedListenerSecurityProtocolMap: "",
		},
	}
	Context("external Access Configuration", func() {

		BeforeEach(func() {
			mockCtrl = gomock.NewController(GinkgoT())
			mockEnv = mocks.NewMockEnvironment(mockCtrl)

			mockEnv.EXPECT().GetNamespace().Return("default").AnyTimes()
			mockEnv.EXPECT().GetExternalIngressPort().Return("9097").AnyTimes()
			mockEnv.EXPECT().GetNodeName().Return("kubelet-0").AnyTimes()
			mockEnv.EXPECT().GetHostName().Return("localhost").AnyTimes()
		})

		AfterEach(func() {
			mockCtrl.Finish()
		})

		for _, test := range tests {
			test := test //necessary to ensure the correct value is passed to the closure
			It(test.name, func() {
				kafkaService := KafkaService{
					Client: testclient.NewSimpleClientset(test.svc, test.node),
					Env:    mockEnv,
				}
				dir, err := ioutil.TempDir("/tmp", "kafka-test")
				defer os.Remove(dir)
				if err != nil {
					log.Fatal(err)
				}
				os.Setenv("LISTENER_SECURITY_PROTOCOL_MAP", "INTERNAL:PLAINTEXT")
				os.Setenv("EXTERNAL_NODEPORT_IP_TYPE", test.nodeportIpType)
				mockEnv.EXPECT().GetNodePortIPType().Return(os.Getenv("EXTERNAL_NODEPORT_IP_TYPE")).AnyTimes()
				err = kafkaService.WriteIngressToPath(dir)
				Expect(err).To(BeNil())

				externalAdvertisedListeners := readFileAsString(fmt.Sprintf("%s/%s", dir, EXTERNAL_ADVERTISED_LISTENERS_PATH)) // just pass the file name
				Expect(externalAdvertisedListeners).To(Equal(test.expectedAdvertisedListeners))

				externalListeners := readFileAsString(fmt.Sprintf("%s/%s", dir, EXTERNAL_LISTENERS)) // just pass the file name
				Expect(externalListeners).To(Equal(test.expectedListeners))

				externalListenerSecurityProtocolMap := readFileAsString(fmt.Sprintf("%s/%s", dir, EXTERNAL_ADVERTISED_LISTENER_SECURITY_MAP)) // just pass the file name
				Expect(externalListenerSecurityProtocolMap).To(Equal(test.expectedListenerSecurityProtocolMap))

				expectedExternalDNS := readFileAsString(fmt.Sprintf("%s/%s", dir, EXTERNAL_DNS)) // just pass the file name
				Expect(expectedExternalDNS).To(Equal(test.expectedExternalDNS))

			})
		}
	})
})

func readFileAsString(path string) string {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return ""
	}
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return ""
	}
	return string(content)
}

func TestService(t *testing.T) {
	RegisterFailHandler(Fail)
	junitReporter := reporters.NewJUnitReporter(fmt.Sprintf("%s-junit.xml", "kafka-utils"))
	RunSpecsWithDefaultAndCustomReporters(t, "KafkaUtils Suite", []Reporter{junitReporter})
}
