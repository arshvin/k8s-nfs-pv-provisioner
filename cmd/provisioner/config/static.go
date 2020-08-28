package config

const (
	/*AnnotationProvisionedBy is annotation used in provisioned PV*/
	AnnotationProvisionedBy = "pv.kubernetes.io/provisioned-by"
	/*AnnotationStorageClass is annotation used in a PVC*/
	AnnotationStorageClass = "volume.beta.kubernetes.io/storage-class"
	/*AnnotationStorageProvisioner is annotation used in a PVC*/
	AnnotationStorageProvisioner = "volume.beta.kubernetes.io/storage-provisioner"

	/*AnnotationOwnerNewAssetUID is the annotation, value of which is able to override the parameter.defaultOwnerAssetUid value of storage class (deprecated)*/
	AnnotationOwnerNewAssetUID = "storage.asset/owner-uid"
	/*AnnotationOwnerNewAssetGID is the annotation, value of which is able to override the parameter.defaultOwnerAssetGid value of storage class (deprecated)*/
	AnnotationOwnerNewAssetGID = "storage.asset/owner-gid"

	/*AnnotationReclaimPolicy is the annotation, value of which is able to override reclaimPolicy value of storage class for particular PVC*/
	AnnotationReclaimPolicy = "volume.pv.provisioner/reclaim-policy"
	/*AnnotationUseExistingAsset is the annotation which allows to reuse existing storage asset if any during provision process for PV*/
	AnnotationUseExistingAsset = "storage-asset.pv.provisioner/reuse-existing"
)
