// Create the security group for mercury instance which allow inboun traffic
// on port 5000 from load balancer inside the vpc.
resource "aws_security_group" "aws_sg_instance_mercury" {
  name = "aws_sg_instance_mercury"
  description = "Security group for mercury"
  vpc_id = aws_vpc.aws_vpc_mercury.id

  ingress {
    description = "Allow traffic inside the vpc"
    from_port = 5000
    to_port = 5000
    protocol = "tcp"
    cidr_blocks = [
      aws_vpc.aws_vpc_mercury.cidr_block]
  }
}

// Create the first mercury instace
resource "aws_instance" "aws_instance_mercury1" {
  ami = data.aws_ami.ubuntu.id
  instance_type = "t3a.micro"
  availability_zone = var.avaiable_zone_1
  subnet_id = aws_subnet.aws_subnet1.id
  key_name = var.key_name
  associate_public_ip_address = true
  vpc_security_group_ids = [
    aws_security_group.aws_sg_mercury_default.id,
    aws_security_group.aws_sg_instance_mercury.id]
  monitoring = true

  tags = {
    Name = "mercury1"
    project = "mercury"
  }

  root_block_device {
    volume_type = "gp2"
    volume_size = 10
  }

  // Create new sudo user `mercury`
  provisioner "remote-exec" {
    inline = [
      "set -x",
      "sudo adduser mercury --gecos \",,,\" --disabled-password",
      "sudo usermod -aG sudo mercury",
      "sudo rsync --archive --chown=mercury:mercury ~/.ssh /home/mercury",
      "sudo bash -c 'echo \"mercury ALL=(ALL) NOPASSWD:ALL\" >> /etc/sudoers'"
    ]

    connection {
      host = coalesce(self.public_ip, self.private_ip)
      type = "ssh"
      user = "ubuntu"
      private_key = file(var.key_file)
    }
  }

  // Upload service file
  provisioner "file" {
    content = local.service_file
    destination = "$HOME/mercury.service"
    connection {
      host = coalesce(self.public_ip, self.private_ip)
      type = "ssh"
      user = "mercury"
      private_key = file(var.key_file)
    }
  }

  // Upload environment file
  provisioner "file" {
    content = local.env_file
    destination = "$HOME/mercury.env"
    connection {
      host = coalesce(self.public_ip, self.private_ip)
      type = "ssh"
      user = "mercury"
      private_key = file(var.key_file)
    }
  }

  // Install goland and build mercury from source code
  provisioner "remote-exec" {
    inline = [
      "set -x",
      "sudo add-apt-repository -y ppa:longsleep/golang-backports",
      "sudo apt update",
      "sudo apt install -y golang-go",
      "git clone https://github.com/renproject/mercury.git",
      "mkdir -p .mercury/bin/",
      "sudo mv mercury.service /etc/systemd/system/",
      "cd mercury/cmd/mercury",
      "go build .",
      "mv mercury ~/.mercury/bin/",
      "sudo service mercury start"
    ]

    connection {
      host = coalesce(self.public_ip, self.private_ip)
      type = "ssh"
      user = "mercury"
      private_key = file(var.key_file)
    }
  }
}

// Create the second mercury instace
resource "aws_instance" "aws_instance_mercury2" {
  ami = data.aws_ami.ubuntu.id
  instance_type = "t3a.micro"
  availability_zone = var.avaiable_zone_2
  subnet_id = aws_subnet.aws_subnet2.id
  key_name = var.key_name
  associate_public_ip_address = true
  vpc_security_group_ids = [
    aws_security_group.aws_sg_mercury_default.id,
    aws_security_group.aws_sg_instance_mercury.id]
  monitoring = true
  tags = {
    Name = "mercury2"
    project = "mercury"
  }

  root_block_device {
    volume_type = "gp2"
    volume_size = 10
  }

  // Create new sudo user `mercury`
  provisioner "remote-exec" {
    inline = [
      "set -x",
      "sudo adduser mercury --gecos \",,,\" --disabled-password",
      "sudo usermod -aG sudo mercury",
      "sudo rsync --archive --chown=mercury:mercury ~/.ssh /home/mercury",
      "sudo bash -c 'echo \"mercury ALL=(ALL) NOPASSWD:ALL\" >> /etc/sudoers'"
    ]

    connection {
      host = coalesce(self.public_ip, self.private_ip)
      type = "ssh"
      user = "ubuntu"
      private_key = file(var.key_file)
    }
  }

  // Upload service file
  provisioner "file" {
    content = local.service_file
    destination = "$HOME/mercury.service"
    connection {
      host = coalesce(self.public_ip, self.private_ip)
      type = "ssh"
      user = "mercury"
      private_key = file(var.key_file)
    }
  }

  // Upload environment file
  provisioner "file" {
    content = local.env_file
    destination = "$HOME/mercury.env"
    connection {
      host = coalesce(self.public_ip, self.private_ip)
      type = "ssh"
      user = "mercury"
      private_key = file(var.key_file)
    }
  }

  // Install goland and build mercury from source code
  provisioner "remote-exec" {
    inline = [
      "set -x",
      "sudo add-apt-repository -y ppa:longsleep/golang-backports",
      "sudo apt update",
      "sudo apt install -y golang-go",
      "git clone https://github.com/renproject/mercury.git",
      "mkdir -p .mercury/bin/",
      "sudo mv mercury.service /etc/systemd/system/",
      "cd mercury/cmd/mercury",
      "go build .",
      "mv mercury ~/.mercury/bin/",
      "sudo service mercury start"
    ]

    connection {
      host = coalesce(self.public_ip, self.private_ip)
      type = "ssh"
      user = "mercury"
      private_key = file(var.key_file)
    }
  }
}