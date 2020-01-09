package config

const (
	/*AnnotationProvisionedBy is annotation used in provisioned PV*/
	AnnotationProvisionedBy = "pv.kubernetes.io/provisioned-by"
	/*AnnotationStorageClass is annotation used in a PVC*/
	AnnotationStorageClass = "volume.beta.kubernetes.io/storage-class"
	/*AnnotationStorageProvisioner is annotation used in a PVC*/
	AnnotationStorageProvisioner = "volume.beta.kubernetes.io/storage-provisioner"
	/*AnnotationOwnerNewAssetUID is a additional parameter in the Storage Class*/
	AnnotationOwnerNewAssetUID = "storage.asset/owner-uid"
	/*AnnotationOwnerNewAssetGID is a additional parameter in the Storage Class*/
	AnnotationOwnerNewAssetGID = "storage.asset/owner-gid"
)
