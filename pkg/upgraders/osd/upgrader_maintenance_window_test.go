package osd

import (
	"fmt"

	"github.com/go-logr/logr"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/types"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	upgradev1alpha1 "github.com/openshift/managed-upgrade-operator/pkg/apis/upgrade/v1alpha1"
	ac "github.com/openshift/managed-upgrade-operator/pkg/availabilitychecks"
	acMocks "github.com/openshift/managed-upgrade-operator/pkg/availabilitychecks/mocks"
	cvMocks "github.com/openshift/managed-upgrade-operator/pkg/clusterversion/mocks"
	"github.com/openshift/managed-upgrade-operator/pkg/drain"
	mockDrain "github.com/openshift/managed-upgrade-operator/pkg/drain/mocks"
	emMocks "github.com/openshift/managed-upgrade-operator/pkg/eventmanager/mocks"
	"github.com/openshift/managed-upgrade-operator/pkg/machinery"
	mockMachinery "github.com/openshift/managed-upgrade-operator/pkg/machinery/mocks"
	mockMaintenance "github.com/openshift/managed-upgrade-operator/pkg/maintenance/mocks"
	mockMetrics "github.com/openshift/managed-upgrade-operator/pkg/metrics/mocks"
	mockScaler "github.com/openshift/managed-upgrade-operator/pkg/scaler/mocks"
	"github.com/openshift/managed-upgrade-operator/util/mocks"
	testStructs "github.com/openshift/managed-upgrade-operator/util/mocks/structs"
)

var _ = Describe("ClusterUpgrader maintenance window tests", func() {
	var (
		logger                   logr.Logger
		upgradeConfigName        types.NamespacedName
		upgradeConfig            *upgradev1alpha1.UpgradeConfig
		mockKubeClient           *mocks.MockClient
		mockCtrl                 *gomock.Controller
		mockMaintClient          *mockMaintenance.MockMaintenance
		mockScaler               *mockScaler.MockScaler
		mockMetricsClient        *mockMetrics.MockMetrics
		mockMachineryClient      *mockMachinery.MockMachinery
		mockCVClient             *cvMocks.MockClusterVersion
		mockDrainStrategyBuilder *mockDrain.MockNodeDrainStrategyBuilder
		mockEMClient             *emMocks.MockEventManager
		mockAC                   *acMocks.MockAvailabilityChecker
		config                   *osdUpgradeConfig
	)

	BeforeEach(func() {
		upgradeConfigName = types.NamespacedName{
			Name:      "test-upgradeconfig",
			Namespace: "test-namespace",
		}
		upgradeConfig = testStructs.NewUpgradeConfigBuilder().WithNamespacedName(upgradeConfigName).GetUpgradeConfig()
		mockCtrl = gomock.NewController(GinkgoT())
		mockKubeClient = mocks.NewMockClient(mockCtrl)
		mockMaintClient = mockMaintenance.NewMockMaintenance(mockCtrl)
		mockMetricsClient = mockMetrics.NewMockMetrics(mockCtrl)
		mockMachineryClient = mockMachinery.NewMockMachinery(mockCtrl)
		mockCVClient = cvMocks.NewMockClusterVersion(mockCtrl)
		mockDrainStrategyBuilder = mockDrain.NewMockNodeDrainStrategyBuilder(mockCtrl)
		mockEMClient = emMocks.NewMockEventManager(mockCtrl)
		mockAC = acMocks.NewMockAvailabilityChecker(mockCtrl)
		logger = logf.Log.WithName("cluster upgrader test logger")
		stepCounter = make(map[upgradev1alpha1.UpgradeConditionType]int)
		config = &osdUpgradeConfig{
			Maintenance: maintenanceConfig{
				ControlPlaneTime: 90,
				IgnoredAlerts: ignoredAlerts{
					ControlPlaneCriticals: []string{"ignoreAlert1SRE", "ignoreAlert2SRE"},
				},
			},
			Scale: scaleConfig{
				TimeOut: 30,
			},
			NodeDrain: drain.NodeDrain{
				ExpectedNodeDrainTime: 8,
			},
		}
	})

	AfterEach(func() {
		mockCtrl.Finish()
	})

	Context("When removing a control plane maintenance window", func() {
		It("Asks the maintenance client to do so", func() {
			mockMaintClient.EXPECT().EndControlPlane()
			result, err := RemoveControlPlaneMaintWindow(mockKubeClient, config, mockScaler, mockDrainStrategyBuilder, mockMetricsClient, mockMaintClient, mockCVClient, mockEMClient, upgradeConfig, mockMachineryClient, []ac.AvailabilityChecker{mockAC}, logger)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(BeTrue())
		})
		It("Indicates when creating the maintenance window has failed", func() {
			mockMaintClient.EXPECT().EndControlPlane().Return(fmt.Errorf("fake error"))
			result, err := RemoveControlPlaneMaintWindow(mockKubeClient, config, mockScaler, mockDrainStrategyBuilder, mockMetricsClient, mockMaintClient, mockCVClient, mockEMClient, upgradeConfig, mockMachineryClient, []ac.AvailabilityChecker{mockAC}, logger)
			Expect(err).To(HaveOccurred())
			Expect(result).To(BeFalse())
		})
	})

	Context("When creating a control plane maintenance window", func() {
		It("Asks the maintenance client to do so", func() {
			mockMaintClient.EXPECT().StartControlPlane(gomock.Any(), upgradeConfig.Spec.Desired.Version, config.Maintenance.IgnoredAlerts.ControlPlaneCriticals)
			result, err := CreateControlPlaneMaintWindow(mockKubeClient, config, mockScaler, mockDrainStrategyBuilder, mockMetricsClient, mockMaintClient, mockCVClient, mockEMClient, upgradeConfig, mockMachineryClient, []ac.AvailabilityChecker{mockAC}, logger)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(BeTrue())
		})
		It("Indicates when creating the maintenance window has failed", func() {
			mockMaintClient.EXPECT().StartControlPlane(gomock.Any(), upgradeConfig.Spec.Desired.Version, config.Maintenance.IgnoredAlerts.ControlPlaneCriticals).Return(fmt.Errorf("fake error"))
			result, err := CreateControlPlaneMaintWindow(mockKubeClient, config, mockScaler, mockDrainStrategyBuilder, mockMetricsClient, mockMaintClient, mockCVClient, mockEMClient, upgradeConfig, mockMachineryClient, []ac.AvailabilityChecker{mockAC}, logger)
			Expect(err).To(HaveOccurred())
			Expect(result).To(BeFalse())
		})
	})

	Context("When creating a worker maintenance window", func() {
		It("Asks the maintenance client to do so", func() {
			mockMachineryClient.EXPECT().IsUpgrading(gomock.Any(), "worker").Return(&machinery.UpgradingResult{IsUpgrading: true, MachineCount: 4, UpdatedCount: 2}, nil)
			mockMaintClient.EXPECT().SetWorker(gomock.Any(), upgradeConfig.Spec.Desired.Version, gomock.Any())
			result, err := CreateWorkerMaintWindow(mockKubeClient, config, mockScaler, mockDrainStrategyBuilder, mockMetricsClient, mockMaintClient, mockCVClient, mockEMClient, upgradeConfig, mockMachineryClient, []ac.AvailabilityChecker{mockAC}, logger)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(BeTrue())
		})
		It("Indicates when creating the maintenance window has failed", func() {
			fakeError := fmt.Errorf("fake error")
			mockMachineryClient.EXPECT().IsUpgrading(gomock.Any(), "worker").Return(&machinery.UpgradingResult{IsUpgrading: true, MachineCount: 4, UpdatedCount: 2}, nil)
			mockMaintClient.EXPECT().SetWorker(gomock.Any(), upgradeConfig.Spec.Desired.Version, gomock.Any()).Return(fakeError)
			result, err := CreateWorkerMaintWindow(mockKubeClient, config, mockScaler, mockDrainStrategyBuilder, mockMetricsClient, mockMaintClient, mockCVClient, mockEMClient, upgradeConfig, mockMachineryClient, []ac.AvailabilityChecker{mockAC}, logger)
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(fakeError))
			Expect(result).To(BeFalse())
		})
		It("Skip creating maintenance window if no pending worker node left", func() {
			mockMachineryClient.EXPECT().IsUpgrading(gomock.Any(), "worker").Return(&machinery.UpgradingResult{IsUpgrading: true, MachineCount: 4, UpdatedCount: 4}, nil)
			result, err := CreateWorkerMaintWindow(mockKubeClient, config, mockScaler, mockDrainStrategyBuilder, mockMetricsClient, mockMaintClient, mockCVClient, mockEMClient, upgradeConfig, mockMachineryClient, []ac.AvailabilityChecker{mockAC}, logger)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(BeTrue())
		})
		It("Does not proceed if isUpgrading check fails", func() {
			mockMachineryClient.EXPECT().IsUpgrading(gomock.Any(), "worker").Return(nil, fmt.Errorf("fake error"))
			result, err := CreateWorkerMaintWindow(mockKubeClient, config, mockScaler, mockDrainStrategyBuilder, mockMetricsClient, mockMaintClient, mockCVClient, mockEMClient, upgradeConfig, mockMachineryClient, []ac.AvailabilityChecker{mockAC}, logger)
			Expect(err).To(HaveOccurred())
			Expect(result).To(BeFalse())
		})
		It("Will not do so if workers are already upgraded", func() {
			mockMachineryClient.EXPECT().IsUpgrading(gomock.Any(), "worker").Return(&machinery.UpgradingResult{IsUpgrading: false}, nil)
			result, err := CreateWorkerMaintWindow(mockKubeClient, config, mockScaler, mockDrainStrategyBuilder, mockMetricsClient, mockMaintClient, mockCVClient, mockEMClient, upgradeConfig, mockMachineryClient, []ac.AvailabilityChecker{mockAC}, logger)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(BeTrue())
		})
	})

	Context("When removing a worker maintenance window", func() {
		It("Asks the maintenance client to do so", func() {
			mockMaintClient.EXPECT().EndWorker()
			result, err := RemoveMaintWindow(mockKubeClient, config, mockScaler, mockDrainStrategyBuilder, mockMetricsClient, mockMaintClient, mockCVClient, mockEMClient, upgradeConfig, mockMachineryClient, []ac.AvailabilityChecker{mockAC}, logger)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(BeTrue())
		})
		It("Indicates when creating the maintenance window has failed", func() {
			mockMaintClient.EXPECT().EndWorker().Return(fmt.Errorf("fake error"))
			result, err := RemoveMaintWindow(mockKubeClient, config, mockScaler, mockDrainStrategyBuilder, mockMetricsClient, mockMaintClient, mockCVClient, mockEMClient, upgradeConfig, mockMachineryClient, []ac.AvailabilityChecker{mockAC}, logger)
			Expect(err).To(HaveOccurred())
			Expect(result).To(BeFalse())
		})
	})
})
