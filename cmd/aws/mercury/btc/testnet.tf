resource "aws_security_group" "aws_security_group_btc_testnet" {
  name = "aws_security_group_btc_testnet"
  description = "Security group for bitcoin testnet node"
  vpc_id = var.vpc_id

  ingress {
    description = "Allow SSH connection "
    from_port = 22
    to_port = 22
    protocol = "tcp"
    cidr_blocks = [
      "0.0.0.0/0"]
  }

  ingress {
    description = "Allow internal jsonrpc request"
    from_port = 18332
    to_port = 18332
    protocol = "tcp"
    cidr_blocks = [
      "10.0.0.0/16"]
  }

  ingress {
    description = "Allow bitcoin nodes communication"
    from_port = 18333
    to_port = 18333
    protocol = "tcp"
    cidr_blocks = [
      "0.0.0.0/0"]
  }

  egress {
    from_port = 0
    to_port = 0
    protocol = "-1"
    cidr_blocks = [
      "0.0.0.0/0"]
  }
}

resource "aws_instance" "bitcoin-testnet-1" {
  ami = data.aws_ami.ubuntu.id
  instance_type = "t3a.medium"
  availability_zone = var.available_zone_1
  key_name = var.key_name
  subnet_id = var.subnet_id_1
  vpc_security_group_ids = [
    aws_security_group.aws_security_group_btc_testnet.id]
  associate_public_ip_address = true
  monitoring = true
  tags = {
    Name = "bitcoin-testnet-1"
  }

  root_block_device {
    volume_type = "gp2"
    volume_size = 50
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
      private_key = file(var.private_key_file)
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
      private_key = file(var.private_key_file)
    }
  }

  // Copy config file
  provisioner "file" {
    content = local.config_file_testnet
    destination = "$HOME/bitcoin.conf"
    connection {
      host = coalesce(self.public_ip, self.private_ip)
      type = "ssh"
      user = "bitcoin"
      private_key = file(var.private_key_file)
    }
  }

  // Install bitcoind and start the service
  provisioner "remote-exec" {
    inline = [
      "set -x",
      "sudo apt-get install --yes software-properties-common",
      "sudo add-apt-repository --yes ppa:luke-jr/bitcoincore",
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
      private_key = file(var.private_key_file)
    }
  }
}

output "btc_testnet_ip" {
  value = aws_instance.bitcoin-testnet-1.private_ip
}