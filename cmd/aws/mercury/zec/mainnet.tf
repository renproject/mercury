resource "aws_security_group" "aws_sg_zec_mainnet" {
  name = "aws_sg_zec_mainnet"
  description = "Security group for zcash mainnet node"
  vpc_id = var.vpc_id

  ingress {
    description = "Allow zcash nodes communication"
    from_port = 8233
    to_port = 8233
    protocol = "tcp"
    cidr_blocks = [
      "0.0.0.0/0"]
  }

  ingress {
    description = "Allow internal jsonrpc request"
    from_port = 8232
    to_port = 8232
    protocol = "tcp"
    cidr_blocks = [
      "10.0.0.0/16"]
  }
}

// First zcash mainnet node instance
resource "aws_instance" "zcash-mainnet-1" {
  ami = var.ami_id
  instance_type = "t3a.large"
  availability_zone = var.available_zone_1
  subnet_id = var.subnet_id_1
  key_name = var.key_name
  associate_public_ip_address = true
  vpc_security_group_ids = [
    var.default_sg_id,
    aws_security_group.aws_sg_zec_mainnet.id]
  monitoring = true
  tags = {
    Name = "zcash-mainnet-1"
    project = "mercury"
  }

  root_block_device {
    volume_type = "gp2"
    volume_size = 50
  }

  // Create new sudo user `zcash`
  provisioner "remote-exec" {
    inline = [
      "set -x",
      "sudo adduser zcash --gecos \",,,\" --disabled-password",
      "sudo usermod -aG sudo zcash",
      "sudo rsync --archive --chown=zcash:zcash ~/.ssh /home/zcash",
      "sudo bash -c 'echo \"zcash ALL=(ALL) NOPASSWD:ALL\" >> /etc/sudoers'"
    ]

    connection {
      host = coalesce(self.public_ip, self.private_ip)
      type = "ssh"
      user = "ubuntu"
      private_key = file(var.key_file)
    }
  }

  // Copy service file
  provisioner "file" {
    content = local.service_file
    destination = "$HOME/zcash.service"
    connection {
      host = coalesce(self.public_ip, self.private_ip)
      type = "ssh"
      user = "zcash"
      private_key = file(var.key_file)
    }
  }

  // Copy config file
  provisioner "file" {
    content = local.config_file_mainnet
    destination = "$HOME/zcash.conf"
    connection {
      host = coalesce(self.public_ip, self.private_ip)
      type = "ssh"
      user = "zcash"
      private_key = file(var.key_file)
    }
  }

  // Install zcashd and start the service
  provisioner "remote-exec" {
    inline = [
      "set -x",
      "sudo apt-get update",
      "sudo apt-get install -y apt-transport-https wget gnupg2",
      "wget -qO - https://apt.z.cash/zcash.asc | sudo apt-key add -",
      "echo \"deb [arch=amd64] https://apt.z.cash/ jessie main\" | sudo tee /etc/apt/sources.list.d/zcash.list",
      "sudo apt-get update",
      "sudo apt-get install -y zcash",
      "zcash-fetch-params",
      "mkdir -p ~/.zcash",
      "mv zcash.conf ./.zcash/",
      "sudo mv zcash.service /etc/systemd/system/zcash.service",
      "sudo service zcash start"
    ]

    connection {
      host = coalesce(self.public_ip, self.private_ip)
      type = "ssh"
      user = "zcash"
      private_key = file(var.key_file)
    }
  }
}

// Second zcash mainnet node instance
resource "aws_instance" "zcash-mainnet-2" {
  ami = var.ami_id
  instance_type = "t3a.large"
  availability_zone = var.available_zone_2
  subnet_id = var.subnet_id_2
  key_name = var.key_name
  associate_public_ip_address = true
  vpc_security_group_ids = [
    var.default_sg_id,
    aws_security_group.aws_sg_zec_mainnet.id]
  monitoring = true
  tags = {
    Name = "zcash-mainnet-2"
    project = "mercury"
  }

  root_block_device {
    volume_type = "gp2"
    volume_size = 50
  }

  // Create new sudo user `zcash`
  provisioner "remote-exec" {
    inline = [
      "set -x",
      "sudo adduser zcash --gecos \",,,\" --disabled-password",
      "sudo usermod -aG sudo zcash",
      "sudo rsync --archive --chown=zcash:zcash ~/.ssh /home/zcash",
      "sudo bash -c 'echo \"zcash ALL=(ALL) NOPASSWD:ALL\" >> /etc/sudoers'"
    ]

    connection {
      host = coalesce(self.public_ip, self.private_ip)
      type = "ssh"
      user = "ubuntu"
      private_key = file(var.key_file)
    }
  }

  // Copy service file
  provisioner "file" {
    content = local.service_file
    destination = "$HOME/zcash.service"
    connection {
      host = coalesce(self.public_ip, self.private_ip)
      type = "ssh"
      user = "zcash"
      private_key = file(var.key_file)
    }
  }

  // Copy config file
  provisioner "file" {
    content = local.config_file_mainnet
    destination = "$HOME/zcash.conf"
    connection {
      host = coalesce(self.public_ip, self.private_ip)
      type = "ssh"
      user = "zcash"
      private_key = file(var.key_file)
    }
  }

  // Install zcashd and start the service
  provisioner "remote-exec" {
    inline = [
      "set -x",
      "sudo apt-get update",
      "sudo apt-get install -y apt-transport-https wget gnupg2",
      "wget -qO - https://apt.z.cash/zcash.asc | sudo apt-key add -",
      "echo \"deb [arch=amd64] https://apt.z.cash/ jessie main\" | sudo tee /etc/apt/sources.list.d/zcash.list",
      "sudo apt-get update",
      "sudo apt-get install -y zcash",
      "zcash-fetch-params",
      "mkdir -p ~/.zcash",
      "mv zcash.conf ./.zcash/",
      "sudo mv zcash.service /etc/systemd/system/zcash.service",
      "sudo service zcash start"
    ]

    connection {
      host = coalesce(self.public_ip, self.private_ip)
      type = "ssh"
      user = "zcash"
      private_key = file(var.key_file)
    }
  }
}

// todo : Cloud watch