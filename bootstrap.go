package main

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	rbac "k8s.io/api/rbac/v1"
	"log"
)

// Init cluster state according to bootstrap.json. Creates namespaces, service accounts, roles, cluster roles, role bindings
// and cluster role bindings. Please note that all errors happened during initialization are ommited(but printed).
// It may be a problem in production application, but in our case we will handle all errors during authorization check.
func InitStateFromBootstrapSettings(settings *BootstrapSettings, clientset *kubernetes.Clientset) {
	createNamespaces(settings, clientset)
	createServiceAccounts(settings, clientset)
	createRoles(settings, clientset)
	createClusterRoles(settings, clientset)
	createRoleBindings(settings, clientset)
	createClusterRoleBindings(settings, clientset)
}

// Removes created namespaces, cluster roles and cluster role bindings
func CleanUp(settings *BootstrapSettings, clientset *kubernetes.Clientset) {
	for _, namespace := range settings.Namespaces {
		clientset.CoreV1().Namespaces().Delete(namespace, &metav1.DeleteOptions{})
	}
	for _, roleBinding := range settings.ClusterRoleBindings {
		clientset.RbacV1().ClusterRoleBindings().Delete(roleBinding.Name, &metav1.DeleteOptions{})
	}
	for _, role := range settings.ClusterRoles {
		clientset.RbacV1().ClusterRoles().Delete(role.Name, &metav1.DeleteOptions{})
	}
}

func createNamespaces(settings *BootstrapSettings, clientset *kubernetes.Clientset) {
	for _, namespace := range settings.Namespaces {
		_, err := clientset.CoreV1().Namespaces().Create(&v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: namespace}})
		if err == nil {
			log.Printf("Created namespace: %s", namespace)
		} else {
			log.Printf("Error occured during namespace '%s' creation. Error: '%s'", namespace, err.Error())
		}
	}
}

func createServiceAccounts(settings *BootstrapSettings, clientset *kubernetes.Clientset) {
	for _, user := range settings.Users {
		_, err := clientset.CoreV1().ServiceAccounts(user.Namespace).Create(&v1.ServiceAccount{
			ObjectMeta: metav1.ObjectMeta{Name: user.Name, Namespace: user.Namespace},
		})
		if err == nil {
			log.Printf("Created user '%s' in namespace '%s'", user.Name, user.Namespace)
		} else {
			log.Printf("Error occured during user creation. User: %s, Namespace: %s, Error: %s",
				user.Name, user.Namespace, err.Error())
		}
	}
}

func createRoles(settings *BootstrapSettings, clientset *kubernetes.Clientset) {
	for _, role := range settings.Roles {
		_, err := clientset.RbacV1().Roles(role.Namespace).Create(buildRole(&role))
		if err == nil {
			log.Printf("Created role '%s' in namespace '%s'", role.Name, role.Namespace)
		} else {
			log.Printf("Error occured during role creation. Name: '%s'. Namespace: '%s'. Error: %s",
				role.Name, role.Namespace, err.Error())
		}
	}
}

func createClusterRoles(settings *BootstrapSettings, clientset *kubernetes.Clientset) {
	for _, role := range settings.ClusterRoles {
		_, err := clientset.RbacV1().ClusterRoles().Create(buildClusterRoles(&role))
		if err == nil {
			log.Printf("Created cluster role '%s'", role.Name)
		} else {
			log.Printf("Error occured during cluster role creation. Name: '%s'. Error: %s",
				role.Name, err.Error())
		}
	}
}

func createRoleBindings(settings *BootstrapSettings, clientset *kubernetes.Clientset) {
	for _, binding := range settings.RoleBindings {
		_, err := clientset.RbacV1().
			RoleBindings(binding.Namespace).
			Create(
			&rbac.RoleBinding{
				ObjectMeta: metav1.ObjectMeta{Name: binding.Name},
				RoleRef: rbac.RoleRef{
					Name:     binding.Role.Name,
					Kind:     binding.Role.Kind,
					APIGroup: binding.Role.ApiGroup,
				},
				Subjects: buildSubjects(binding.Subjects),
			})
		if err == nil {
			log.Printf("Created role binding '%s' in namespace '%s'", binding.Name, binding.Namespace)
		} else {
			log.Printf("Couldn't create role binding '%s' in namespace '%s'. Error: %s",
				binding.Name, binding.Namespace, err)
		}

	}
}

func createClusterRoleBindings(settings *BootstrapSettings, clientset *kubernetes.Clientset) {
	for _, binding := range settings.ClusterRoleBindings {
		_, err := clientset.RbacV1().
			ClusterRoleBindings().
			Create(
			&rbac.ClusterRoleBinding{
				ObjectMeta: metav1.ObjectMeta{Name: binding.Name},
				RoleRef:    rbac.RoleRef{Name: binding.Role.Name, Kind: binding.Role.Kind, APIGroup: binding.Role.ApiGroup},
				Subjects:   buildSubjects(binding.Subjects),
			})
		if err == nil {
			log.Printf("Created cluster role binding '%s'", binding.Name)
		} else {
			log.Printf("Error occured during cluster role binding creation. Name: '%s'. Error: '%s'", binding.Name, err.Error())
		}
	}
}

func buildClusterRoles(settings *ClusterRoleSettings) *rbac.ClusterRole {
	return &rbac.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{Name: settings.Name},
		Rules:      buildRules(settings.Rules),
	}
}

func buildRole(settings *RoleSettings) *rbac.Role {
	return &rbac.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      settings.Name,
			Namespace: settings.Namespace},
		Rules: buildRules(settings.Rules)}
}

func buildSubjects(subjects []SubjectSettings) []rbac.Subject {
	out := make([]rbac.Subject, len(subjects))
	for i, subject := range subjects {
		out[i] = rbac.Subject{Kind: subject.Kind,
			Name: subject.Name,
			APIGroup: subject.ApiGroup,
			Namespace: subject.Namespace,
		}
	}
	return out
}

func buildRules(input []Rule) []rbac.PolicyRule {
	out := make([]rbac.PolicyRule, len(input))
	for i, rule := range input {
		out[i] = rbac.PolicyRule{
			Verbs:     rule.Verbs,
			Resources: rule.Resources,
			APIGroups: rule.ApiGroups,
		}
	}
	return out
}
