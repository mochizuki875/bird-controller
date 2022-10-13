/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"fmt"
	"math"

	"github.com/go-logr/logr"
	"github.com/google/uuid"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	birdv1 "example.com/bird-controller/api/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// BirdReconciler reconciles a Bird object
type BirdReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=bird.my.domain,resources=birds,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=bird.my.domain,resources=birds/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=bird.my.domain,resources=birds/finalizers,verbs=update
//+kubebuilder:rbac:groups=bird.my.domain,resources=eggs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=bird.my.domain,resources=eggs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=bird.my.domain,resources=eggs/finalizers,verbs=update

func (r *BirdReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	defer fmt.Println("=== Finish Reconcile ===")
	fmt.Println("=== Start Reconcile ===")

	var bird birdv1.Bird
	var eggList birdv1.EggList
	var err error

	// ①更新のあったBirdリソースを取得(from cache)
	if err = r.Get(ctx, req.NamespacedName, &bird); err != nil {
		log.Error(err, "Faild to fetch bird.")
		return ctrl.Result{}, client.IgnoreNotFound(err) // 取得に失敗したらエラーをスキップする
	}

	// ②Birdの管理するEgg一覧をListで取得する
	// Namespaceとindex(SetupWithManagerで付与)を元にcache上のEggリソースをフィルターして取得する
	if err = r.List(ctx, &eggList, client.InNamespace(req.Namespace), client.MatchingFields{eggOwnerKey: bird.Name}); err != nil {
		log.Error(err, "Faild to list eggs.")
		return ctrl.Result{}, err
	}

	// ③BirdリソースからEggNumbersを取得
	eggNumber := *bird.Spec.EggNumbers

	// EggNumberと取得したEgg数を比較
	diff := int32(len(eggList.Items)) - eggNumber

	// (EggNumbers < Egg数) 過剰なEggリソースを削除
	if diff > 0 {
		// Debug
		fmt.Println("Delete ", diff, " Eggs.")

		// eggListから削除対象のSliceを取得
		excessiveEggList := eggList.Items[eggNumber:]

		for _, egg := range excessiveEggList {
			log.Info("Delete Egg: " + egg.Name)
			// log.Info(egg.Name + "を叩き割る。")
			if err = r.Delete(ctx, &egg); err != nil {
				log.Error(err, "Faild to delete Egg")
				return ctrl.Result{}, err
			}
		}

	}
	// (EggNumbers > Egg数) EggNumbers分のEggリソースを作成(CreateOrUpdate)
	if diff < 0 {
		// 不足分のEggリソースを作成
		diff = int32(math.Abs(float64(diff)))

		fmt.Println("Create ", diff, " Eggs.")

		for i := 0; i < int(diff); i++ {
			// BirdリソースをもとにEggリソースを作成するメソッド
			if err = r.CreateEgg(ctx, log, &bird); err != nil {
				return ctrl.Result{}, err
			}
		}
	}

	// ④Statusを更新
	eggNumberFlag := false

	// Birdリソースが管理するEgg一覧を取得
	if err = r.List(ctx, &eggList, client.InNamespace(req.Namespace), client.MatchingFields{eggOwnerKey: bird.Name}); err != nil {
		log.Error(err, "Faild to list eggs.")
		return ctrl.Result{}, err
	}

	eggListLen := int32(0)
	eggListLen = int32(len(eggList.Items))

	statusEggnum := int32(0)
	if bird.Status.EggNumbers != nil {
		statusEggnum = *bird.Status.EggNumbers
	}

	// Birdリソースの.Status.EggNumbersとeggListLenの差分比較を行い、差分があれば.Status.EggNumbersを更新
	if statusEggnum != eggListLen {
		bird.Status.EggNumbers = &eggListLen
		eggNumberFlag = true
	}

	if eggNumberFlag {
		log.Info("Update Bird Status: " + bird.Name)
		if err = r.Status().Update(ctx, &bird); err != nil {
			log.Error(err, "Unable to update Bird")
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

var (
	eggOwnerKey = ".metadata.controller"
	apiGVStr    = birdv1.GroupVersion.String()
)

// SetupWithManager sets up the controller with the Manager.

func (r *BirdReconciler) SetupWithManager(mgr ctrl.Manager) error {

	// cache上のEggリソースにOwnerReferenceに基づくIndexを付与
	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &birdv1.Egg{}, eggOwnerKey, IndexByOwner); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&birdv1.Bird{}).
		Owns(&birdv1.Egg{}). // このリソースに変更があったらOwnリソースをreconcileする
		Complete(r)
}

// OwnerReferenceの付与状況を確認し、Indexとして付与する値を決める関数
func IndexByOwner(rawObj client.Object) []string {
	/*
	  アサーションによりclient.Object型として渡されたrawObjが*appsv1.Deployment型であることを確認しつつ、
	  *appsv1.Deployment型として値を取得
	  https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/client#Object
	*/
	egg := rawObj.(*birdv1.Egg)

	// OwnerReferenceへのポインタを取得
	// https://pkg.go.dev/k8s.io/apimachinery@v0.25.0/pkg/apis/meta/v1#OwnerReference
	owner := metav1.GetControllerOf(egg)

	// OwnerReferenceが設定されていない場合はスキップ
	if owner == nil {
		return nil
	}

	// OwnerReferenceのapiVersionまたはKindがBirdと一致しない場合はスキップ
	if owner.APIVersion != apiGVStr || owner.Kind != "Bird" {
		return nil
	}

	return []string{owner.Name} // .metadata.controller: owner.NameというIndexを追加

}

func (r *BirdReconciler) CreateEgg(ctx context.Context, log logr.Logger, bird *birdv1.Bird) error {

	// UUIDオブジェクトを生成
	uuidobj, err := uuid.NewUUID()
	if err != nil {
		log.Error(err, "Faild to create UUID for egg.")
		return err
	}
	UUID := uuidobj.String()
	Labels := map[string]string{
		"egg-uuid": UUID,
	}

	// Eggを作成
	egg := &birdv1.Egg{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "egg-from-" + bird.Name + "-" + UUID,
			Namespace: bird.Namespace,
			Labels:    Labels,
		},
	}

	log.Info("Create Egg: " + egg.Name)
	// log.Info(egg.Name + "を産卵しました！")

	if _, err := ctrl.CreateOrUpdate(ctx, r.Client, egg, func() error {

		// egg.Spec.Parentを設定
		if egg.Spec.Parent == "" {
			egg.Spec.Parent = bird.Name
		}

		// OwnerReferenceを設定
		if err = ctrl.SetControllerReference(bird, egg, r.Scheme); err != nil {
			log.Error(err, "Unable to set OwnerReference from Bird to Egg")
		}

		return nil

	}); err != nil {
		log.Error(err, "Unable to ensure Egg is correct")
		return err
	}

	// EggのStatus更新
	egg.Status.Parent = egg.Spec.Parent
	if err = r.Status().Update(ctx, egg); err != nil {
		log.Error(err, "Unable to update Egg")
		return err
	}

	return nil

}
