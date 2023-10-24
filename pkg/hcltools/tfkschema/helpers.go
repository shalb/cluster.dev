package tfkschema

import (
	"reflect"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type K8sObject struct {
	Object           runtime.Object
	GroupVersionKind *schema.GroupVersionKind
}

func ObjectMeta(obj runtime.Object) metav1.ObjectMeta {
	v := reflect.ValueOf(obj)

	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	metaF := v.FieldByName("ObjectMeta")

	return metaF.Interface().(metav1.ObjectMeta)
}

func TypeMeta(obj runtime.Object) metav1.TypeMeta {
	v := reflect.ValueOf(obj)

	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	metaF := v.FieldByName("TypeMeta")

	return metaF.Interface().(metav1.TypeMeta)
}
