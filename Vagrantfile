# -*- mode: ruby -*-
# vi: set ft=ruby :

Vagrant.configure("2") do |config|
    config.vm.define "microk8s" do |node|
      node.vm.box = "bento/centos-7.6"
      node.vm.box_check_update = false
  
      node.vm.hostname = "microk8s"
      vagrant_share = '/vagrant'
      node.vm.synced_folder '.', vagrant_share
  
      node.vm.provider :virtualbox do |vb|
        vb.memory = 2048
        vb.cpus = 2
  
        # disable usb 2.0 and audio support
        vb.customize ["modifyvm", :id, "--usb", "off"]
        vb.customize ["modifyvm", :id, "--usbehci", "off"]
        vb.customize ["modifyvm", :id, "--audio", "none"]
        vb.customize ["modifyvm", :id, "--natdnshostresolver1", "on"]
        vb.customize ["modifyvm", :id, "--natdnsproxy1", "on"]
      end
  
      node.vm.provision "shell", path: "./vm_provisioning/provisioning.sh", args: vagrant_share, keep_color: true
    end
  end
  