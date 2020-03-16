resource "aws_security_group" "aws_sg_bch_mainnet" {
  name = "aws_sg_bch_mainnet"
  description = "Security group for bitcoin cash mainnet node"
  vpc_id = var.vpc_id

  ingress {
    description = "Allow bitcoin cash nodes communication"
    from_port = 8333
    to_port = 8333
    protocol = "tcp"
    cidr_blocks = [
      "0.0.0.0/0"]
  }

  ingress {
    description = "Allow internal jsonrpc request"
    from_port = 8332
    to_port = 8332
    protocol = "tcp"
    cidr_blocks = [
      "10.0.0.0/16"]
  }
}

// First bitcoin cash node instance
resource "aws_instance" "bcash-mainnet-1" {
  ami = var.ami_id
  instance_type = "t3a.large"
  availability_zone = var.available_zone_1
  subnet_id = var.subnet_id_1
  key_name = var.key_name
  associate_public_ip_address = true
  vpc_security_group_ids = [
    var.default_sg_id,
    aws_security_group.aws_sg_bch_mainnet.id]
  monitoring = true
  tags = {
    Name = "bcash-mainnet-1"
    project = "mercury"
  }

  root_block_device {
    volume_type = "gp2"
    volume_size = 200
  }

  // Create new sudo user `bitcoin`
  provisioner "remote-exec" {
    inline = [
      "set -x",
      "sudo adduser bitcoin --gecos \",,,\" --disabled-password",
      "sudo usermod -aG sudo bitcoin",
      "sudo rsync --archive --chown=bitcoin:bitcoin ~/.ssh /home/bitcoin",
      "sudo bash -c 'echo \"bitcoin ALL=(ALL) NOPASSWD:ALL\" >> /etc/sudoers'"
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
    destination = "$HOME/bitcoin.service"
    connection {
      host = coalesce(self.public_ip, self.private_ip)
      type = "ssh"
      user = "bitcoin"
      private_key = file(var.key_file)
    }
  }

  // Copy config file
  provisioner "file" {
    content = local.config_file_mainnet
    destination = "$HOME/bitcoin.conf"
    connection {
      host = coalesce(self.public_ip, self.private_ip)
      type = "ssh"
      user = "bitcoin"
      private_key = file(var.key_file)
    }
  }

  // Install bitcoind and start the service
  provisioner "remote-exec" {
    inline = [
      "set -x",
      "sudo apt-get update",
      "sudo apt-get install --yes software-properties-common",
      "sudo add-apt-repository --yes ppa:bitcoin-abc/ppa",
      "sudo apt-get update",
      "sudo apt-get install --yes bitcoind",
      "mkdir ~/.bitcoin",
      "mv bitcoin.conf ./.bitcoin/",
      "sudo mv bitcoin.service /etc/systemd/system/bitcoin.service",
      "sudo service bitcoin start"
    ]

    connection {
      host = coalesce(self.public_ip, self.private_ip)
      type = "ssh"
      user = "bitcoin"
      private_key = file(var.key_file)
    }
  }
}

resource "aws_instance" "bcash-mainnet-2" {
  ami = var.ami_id
  instance_type = "t3a.large"
  availability_zone = var.available_zone_2
  subnet_id = var.subnet_id_2
  key_name = var.key_name
  associate_public_ip_address = true
  vpc_security_group_ids = [
    var.default_sg_id,
    aws_security_group.aws_sg_bch_mainnet.id]
  monitoring = true
  tags = {
    Name = "bcash-mainnet-2"
    project = "mercury"
  }

  root_block_device {
    volume_type = "gp2"
    volume_size = 200
  }

  // Create new sudo user `bitcoin`
  provisioner "remote-exec" {
    inline = [
      "set -x",
      "sudo adduser bitcoin --gecos \",,,\" --disabled-password",
      "sudo usermod -aG sudo bitcoin",
      "sudo rsync --archive --chown=bitcoin:bitcoin ~/.ssh /home/bitcoin",
      "sudo bash -c 'echo \"bitcoin ALL=(ALL) NOPASSWD:ALL\" >> /etc/sudoers'"
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
    destination = "$HOME/bitcoin.service"
    connection {
      host = coalesce(self.public_ip, self.private_ip)
      type = "ssh"
      user = "bitcoin"
      private_key = file(var.key_file)
    }
  }

  // Copy config file
  provisioner "file" {
    content = local.config_file_mainnet
    destination = "$HOME/bitcoin.conf"
    connection {
      host = coalesce(self.public_ip, self.private_ip)
      type = "ssh"
      user = "bitcoin"
      private_key = file(var.key_file)
    }
  }

  // Install bitcoind and start the service
  provisioner "remote-exec" {
    inline = [
      "set -x",
      "sudo apt-get update",
      "sudo apt-get install --yes software-properties-common",
      "sudo add-apt-repository --yes ppa:bitcoin-abc/ppa",
      "sudo apt-get update",
      "sudo apt-get install --yes bitcoind",
      "mkdir ~/.bitcoin",
      "mv bitcoin.conf ./.bitcoin/",
      "sudo mv bitcoin.service /etc/systemd/system/bitcoin.service",
      "sudo service bitcoin start"
    ]

    connection {
      host = coalesce(self.public_ip, self.private_ip)
      type = "ssh"
      user = "bitcoin"
      private_key = file(var.key_file)
    }
  }
}

// todo : Cloud watch