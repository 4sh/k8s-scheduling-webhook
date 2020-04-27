package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/gorilla/pat"
	"html"
	"net/http"
	"os"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	whhttp "github.com/slok/kubewebhook/pkg/http"
	"github.com/slok/kubewebhook/pkg/log"
	mutatingwh "github.com/slok/kubewebhook/pkg/webhook/mutating"
)

func handleRoot(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "hello %q", html.EscapeString(r.URL.Path))
}

func AppendNodeSelector(a *corev1.NodeAffinity, s corev1.NodeSelectorRequirement) {
	if a.RequiredDuringSchedulingIgnoredDuringExecution == nil {
		a.RequiredDuringSchedulingIgnoredDuringExecution = &corev1.NodeSelector{}
	}

	a.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms = append(a.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms, corev1.NodeSelectorTerm{
		MatchExpressions: []corev1.NodeSelectorRequirement{s},
	})
}

func AppendNodeAffinitySelector(nodeSelection *corev1.PodSpec, s corev1.NodeSelectorRequirement) {
	if nodeSelection.Affinity == nil {
		nodeSelection.Affinity = &corev1.Affinity{}
	}
	if nodeSelection.Affinity.NodeAffinity == nil {
		nodeSelection.Affinity.NodeAffinity = &corev1.NodeAffinity{}
	}
	AppendNodeSelector(nodeSelection.Affinity.NodeAffinity, s)
}

func annotatePodMutator(ctx context.Context, obj metav1.Object) (bool, error) {
	pod, ok := obj.(*corev1.Pod)
	if !ok {
		// If not a pod just continue the mutation chain(if there is one) and don't do nothing.
		return false, nil
	}

	// Mutate our object with the required annotations.
	if pod.Annotations == nil {
		pod.Annotations = make(map[string]string)
	}
	pod.Annotations["mutated"] = "true"
	pod.Annotations["mutator"] = fmt.Sprintf("scheduling-%s", ctx.Value("id"))

	AppendNodeAffinitySelector(&pod.Spec, corev1.NodeSelectorRequirement{
		Key:      "autoscaling",
		Operator: "In",
		Values:   []string{"true"},
	})

	return false, nil
}

type config struct {
	certFile string
	keyFile  string
}

func initFlags() *config {
	cfg := &config{}

	fl := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	fl.StringVar(&cfg.certFile, "tls-cert-file", "", "TLS certificate file")
	fl.StringVar(&cfg.keyFile, "tls-key-file", "", "TLS key file")

	fl.Parse(os.Args[1:])
	return cfg
}

func HandlerWithIdFor(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get(":id")
		handler.ServeHTTP(w, r.Clone(context.WithValue(r.Context(), "id", id)))
	})
}

func main() {
	logger := &log.Std{Debug: true}

	cfg := initFlags()

	// Create our mutator
	mt := mutatingwh.MutatorFunc(annotatePodMutator)

	mcfg := mutatingwh.WebhookConfig{
		Name: "podAnnotate",
		Obj:  &corev1.Pod{},
	}
	wh, err := mutatingwh.NewWebhook(mcfg, mt, nil, nil, logger)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error creating webhook: %s", err)
		os.Exit(1)
	}

	// Get the handler for our webhook.
	whHandler, err := whhttp.HandlerFor(wh)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error creating webhook handler: %s", err)
		os.Exit(1)
	}

	var mux = pat.New()

	mux.Get("/", handleRoot)
	mux.Add("POST", "/mutate/{id}", HandlerWithIdFor(whHandler))

	logger.Infof("Listening on :8080")
	s := &http.Server{
		Addr:           ":8443",
		Handler:        mux,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1048576
	}

	err = s.ListenAndServeTLS(cfg.certFile, cfg.keyFile)

	if err != nil {
		fmt.Fprintf(os.Stderr, "error serving webhook: %s", err)
		os.Exit(1)
	}
}
