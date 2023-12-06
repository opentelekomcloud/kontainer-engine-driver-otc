package opentelekomcloud

import (
	"context"
	"fmt"
	"time"

	errs "github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	rbacV1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	cattleSecretName          = "cattle-secret"
	cattleNamespace           = "cattle-system"
	clusterAdmin              = "cluster-admin"
	kontainerEngine           = "kontainer-engine"
	newClusterRoleBindingName = "system-netes-default-clusterRoleBinding"
)

// GenerateServiceAccountToken generate a serviceAccountToken for clusterAdmin given a rest clientset
func generateServiceAccountToken(clientset kubernetes.Interface) (string, error) {
	_, err := clientset.CoreV1().Namespaces().Create(context.TODO(), &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: cattleNamespace,
		},
	}, metav1.CreateOptions{})
	if err != nil && !errors.IsAlreadyExists(err) {
		return "", err
	}

	serviceAccount := &v1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name: kontainerEngine,
		},
		Secrets: []v1.ObjectReference{
			{
				Kind:      "Secret",
				Namespace: cattleNamespace,
				Name:      cattleSecretName,
			},
		},
	}

	clusterServAcc, err := clientset.CoreV1().ServiceAccounts(cattleNamespace).Get(context.TODO(), kontainerEngine, metav1.GetOptions{})
	switch {
	case errors.IsNotFound(err):
		clusterServAcc, err = clientset.CoreV1().ServiceAccounts(cattleNamespace).Create(context.TODO(), serviceAccount, metav1.CreateOptions{})
		if err != nil {
			return "", fmt.Errorf("error creating service account: %v", err)
		}
	case err != nil:
		return "", fmt.Errorf("error getting service account: %v", err)
	}

	adminRole := &rbacV1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: clusterAdmin,
		},
		Rules: []rbacV1.PolicyRule{
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
	clusterAdminRole, err := clientset.RbacV1().ClusterRoles().Get(context.TODO(), clusterAdmin, metav1.GetOptions{})
	if err != nil {
		clusterAdminRole, err = clientset.RbacV1().ClusterRoles().Create(context.TODO(), adminRole, metav1.CreateOptions{})
		if err != nil {
			return "", fmt.Errorf("error creating admin role: %v", err)
		}
	}

	clusterRoleBinding := &rbacV1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: newClusterRoleBindingName,
		},
		Subjects: []rbacV1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      clusterServAcc.Name,
				Namespace: cattleNamespace,
			},
		},
		RoleRef: rbacV1.RoleRef{
			Kind:     "ClusterRole",
			Name:     clusterAdminRole.Name,
			APIGroup: rbacV1.GroupName,
		},
	}
	if _, err = clientset.RbacV1().ClusterRoleBindings().Create(context.TODO(), clusterRoleBinding, metav1.CreateOptions{}); err != nil && !errors.IsAlreadyExists(err) {
		return "", fmt.Errorf("error creating role bindings: %v", err)
	}

	tokenSecret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cattleSecretName,
			Namespace: cattleNamespace,
			Annotations: map[string]string{
				"kubernetes.io/service-account.name": kontainerEngine,
				"kubernetes.io/service-account.uid":  string(clusterServAcc.UID),
			},
		},
		Type: v1.SecretTypeServiceAccountToken,
	}

	if _, err = clientset.CoreV1().Secrets(cattleNamespace).Create(context.TODO(), tokenSecret, metav1.CreateOptions{}); err != nil && !errors.IsAlreadyExists(err) {
		return "", fmt.Errorf("error creating service secret: %v", err)
	}

	start := time.Millisecond * 250
	for i := 0; i < 5; i++ {
		time.Sleep(start)

		if clusterServAcc, err = clientset.CoreV1().ServiceAccounts(cattleNamespace).Get(context.TODO(), clusterServAcc.Name, metav1.GetOptions{}); err != nil {
			return "", fmt.Errorf("error getting service account: %v", err)
		}

		if len(clusterServAcc.Secrets) > 0 {
			secretObj, err := clientset.CoreV1().Secrets(cattleNamespace).Get(context.TODO(), cattleSecretName, metav1.GetOptions{})
			if err != nil {
				return "", fmt.Errorf("error getting secret: %v", err)
			}
			if token, ok := secretObj.Data["token"]; ok {
				return string(token), nil
			}
		}
		start *= 2
	}

	return "", errs.New("failed to fetch token")
}
