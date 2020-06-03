package opentelekomcloud

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/rancher/kontainer-engine/drivers/options"
	"github.com/rancher/kontainer-engine/types"
	v1 "k8s.io/api/core/v1"
	"k8s.io/api/rbac/v1beta1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	clusterAdmin     = "cluster-admin"
	netesDefault     = "netes-default"
	defaultNamespace = "cattle-system"
)

var get = options.GetValueFromDriverOptions

type strFromOpts func(keys ...string) string
type strSliceFromOpts func(keys ...string) []string
type intFromOpts func(keys ...string) int64
type boolFromOpts func(keys ...string) bool

// Produce options getters for each argument type
func getters(opts *types.DriverOptions) (strFromOpts, strSliceFromOpts, intFromOpts, boolFromOpts) {
	s := func(k ...string) string {
		return get(opts, types.StringType, k...).(string)
	}
	sl := func(k ...string) []string {
		return get(opts, types.StringSliceType, k...).(*types.StringSlice).Value
	}
	i := func(k ...string) int64 {
		return get(opts, types.IntType, k...).(int64)
	}
	b := func(k ...string) bool {
		return get(opts, types.BoolType, k...).(bool)
	}
	return s, sl, i, b
}

func generateServiceAccountToken(ctx context.Context, clientSet kubernetes.Interface) (string, error) {
	serviceAccount := &v1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name: netesDefault,
		},
	}

	_, err := clientSet.CoreV1().ServiceAccounts(defaultNamespace).Create(ctx, serviceAccount, metav1.CreateOptions{})
	if err != nil && !apierrors.IsAlreadyExists(err) {
		return "", fmt.Errorf("error creating service account: %v", err)
	}

	adminRole := &v1beta1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: clusterAdmin,
		},
		Rules: []v1beta1.PolicyRule{
			{
				APIGroups: []string{"*"},
				Resources: []string{"*"},
				Verbs:     []string{"*"},
			},
			{
				NonResourceURLs: []string{"*"},
				Verbs:           []string{"*"},
			},
		},
	}
	clusterAdminRole, err := clientSet.RbacV1beta1().ClusterRoles().Get(ctx, clusterAdmin, metav1.GetOptions{})
	if err != nil {
		clusterAdminRole, err = clientSet.RbacV1beta1().ClusterRoles().Create(ctx, adminRole, metav1.CreateOptions{})
		if err != nil {
			return "", fmt.Errorf("error creating admin role: %v", err)
		}
	}

	clusterRoleBinding := &v1beta1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: "netes-default-clusterRoleBinding",
		},
		Subjects: []v1beta1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      serviceAccount.Name,
				Namespace: "default",
				APIGroup:  v1.GroupName,
			},
		},
		RoleRef: v1beta1.RoleRef{
			Kind:     "ClusterRole",
			Name:     clusterAdminRole.Name,
			APIGroup: v1beta1.GroupName,
		},
	}
	if _, err = clientSet.RbacV1beta1().ClusterRoleBindings().Create(ctx, clusterRoleBinding, metav1.CreateOptions{}); err != nil && !apierrors.IsAlreadyExists(err) {
		return "", fmt.Errorf("error creating role bindings: %v", err)
	}

	start := time.Millisecond * 250
	for i := 0; i < 5; i++ {
		time.Sleep(start)
		if serviceAccount, err = clientSet.CoreV1().ServiceAccounts(defaultNamespace).Get(ctx, serviceAccount.Name, metav1.GetOptions{}); err != nil {
			return "", fmt.Errorf("error getting service account: %v", err)
		}

		if len(serviceAccount.Secrets) > 0 {
			secret := serviceAccount.Secrets[0]
			secretObj, err := clientSet.CoreV1().Secrets(defaultNamespace).Get(ctx, secret.Name, metav1.GetOptions{})
			if err != nil {
				return "", fmt.Errorf("error getting secret: %v", err)
			}
			if token, ok := secretObj.Data["token"]; ok {
				return string(token), nil
			}
		}
		start = start * 2
	}

	return "", errors.New("failed to fetch serviceAccountToken")
}
