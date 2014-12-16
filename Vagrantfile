Vagrant.configure('2') do |config|
  config.vm.box = 'cloudfoundry/bosh-lite'
  config.vm.box_version = '388'

  [:virtualbox, :vmware_fusion, :vmware_desktop, :vmware_workstation].each do |provider|
    config.vm.provider provider do |v, override|
      v.memory = 1024 * 4
      v.cpus = 4
    end
  end

  config.vm.provider :aws do |v, override|
    v.tags = {
      'PipelineName' => 'bosh-micro-cli'
    }
  end

  config.vm.synced_folder Dir.pwd, '/vagrant', disabled: true
  config.vm.provision :shell, inline: "mkdir -p /vagrant && chmod 777 /vagrant"
  config.vm.provision :shell, inline: "chmod 777 /var/vcap/sys/log/cpi"
end
