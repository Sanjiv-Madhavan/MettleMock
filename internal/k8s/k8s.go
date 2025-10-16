package k8s

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"reflect"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// UpsertOpaqueSecret creates the Secret if missing or updates its .data if it exists.
// - Type is always Opaque (we never mutate type on update to avoid immutability errors).
// - Sets an owner reference to `owner` so GC cleans it up with the CR.
// - Adds a checksum annotation for easy drift detection in diffs.
func UpsertOpaqueSecret(
	ctx context.Context,
	c client.Client,
	namespace, name string,
	data map[string][]byte,
	owner client.Object,
) error {
	sum := checksumData(data)

	// Try GET first.
	var s corev1.Secret
	key := client.ObjectKey{Namespace: namespace, Name: name}
	err := c.Get(ctx, key, &s)
	if err != nil {
		if !apierrors.IsNotFound(err) {
			return err
		}
		// Create path
		s = corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
				Annotations: map[string]string{
					"checksum/data": sum,
				},
			},
			Type: corev1.SecretTypeOpaque,
			Data: data,
		}
		if owner != nil {
			// Set owner so Secret is garbage-collected when CR is deleted
			if err := controllerutil.SetControllerReference(owner, &s, c.Scheme()); err != nil {
				return err
			}
		}
		return c.Create(ctx, &s)
	}

	// Update path (avoid changing immutable fields like .type)
	// Compare data to skip no-op update
	if reflect.DeepEqual(s.Data, data) && s.Annotations["checksum/data"] == sum {
		return nil // no change
	}

	// Patch with retry for optimistic concurrency conflicts
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		if err := c.Get(ctx, key, &s); err != nil {
			return err
		}
		if s.Annotations == nil {
			s.Annotations = map[string]string{}
		}
		s.Annotations["checksum/data"] = sum
		s.Data = data
		if owner != nil {
			if err := controllerutil.SetControllerReference(owner, &s, c.Scheme()); err != nil {
				return err
			}
		}
		return c.Update(ctx, &s)
	})
}

// checksumData is a small helper to annotate the Secret with a stable hash of its data.
func checksumData(data map[string][]byte) string {
	h := sha256.New()
	// Deterministic hashing: iterate keys in lexical order
	keys := make([]string, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}
	// simple insertion sort to avoid extra deps
	for i := 1; i < len(keys); i++ {
		for j := i; j > 0 && keys[j] < keys[j-1]; j-- {
			keys[j], keys[j-1] = keys[j-1], keys[j]
		}
	}
	for _, k := range keys {
		h.Write([]byte(k))
		h.Write([]byte{0})
		h.Write(data[k])
		h.Write([]byte{0})
	}
	return hex.EncodeToString(h.Sum(nil))
}
