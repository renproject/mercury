resource "aws_security_group" "aws_security_group_zec_mainnet_testnet" {
  name = "aws_security_group_zec_mainnet_testnet"
  description = "Security group for zcash testnet node"
  vpc_id = var.vpc_id

  ingress {
    description = "Allow SSH connection "
    from_port = 22
    to_port = 22
    protocol = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    description = "Allow internal jsonrpc request"
    from_port   = 18232
    to_port     = 18232
    protocol    = "tcp"
    cidr_blocks = ["10.0.0.0/16"]
  }

  ingress {
    description = "Allow zcash nodes communication"
    from_port = 18233
    to_port = 18233
    protocol = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port = 0
    to_port = 0
    protocol = "-1"
    cidr_blocks = [
      "0.0.0.0/0"]
  }
}

resource "aws_instance" "zcash-testnet-1" {
  ami = data.aws_ami.ubuntu.id
  instance_type = "t3a.medium"
  availability_zone = var.available_zone_1
  key_name = var.key_name
  subnet_id = var.subnet_id_1
  vpc_security_group_ids  = [aws_security_group.aws_security_group_zec_mainnet_testnet.id]
  associate_public_ip_address = true
  monitoring = true
  tags = {
    Name = "zcash-testnet-1"
  }

  root_block_device {
    volume_type = "gp2"
    volume_size = 30
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
      private_key = file(var.private_key_file)
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
      private_key = file(var.private_key_file)
    }
  }

  // Copy config file
  provisioner "file" {
    content = local.config_file_testnet
    destination = "$HOME/zcash.conf"
    connection {
      host = coalesce(self.public_ip, self.private_ip)
      type = "ssh"
      user = "zcash"
      private_key = file(var.private_key_file)
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
      private_key = file(var.private_key_file)
    }
  }
}

output "zec_testnet_ip"{
  value = aws_instance.zcash-testnet-1.private_ip
}