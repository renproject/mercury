// Create a new VPC
resource "aws_vpc" "aws_vpc_mercury" {
  cidr_block = "10.0.0.0/16"
  instance_tenancy = "default"

  tags = {
    project = "mercury"
  }
}

// Create a gateway for the VPC
resource "aws_internet_gateway" "aws_internet_gateway_mercury" {
  vpc_id = aws_vpc.aws_vpc_mercury.id

  tags = {
    project = "mercury"
  }
}

// Create a route table for all the public subnets
resource "aws_route_table" "aws_route_table_mercury" {
  vpc_id = aws_vpc.aws_vpc_mercury.id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.aws_internet_gateway_mercury.id
  }

  tags = {
    project = "mercury"
  }
}

// Create subnets for each available zone
resource "aws_subnet" "aws_subnet1" {
  vpc_id = aws_vpc.aws_vpc_mercury.id
  cidr_block = "10.0.1.0/24"
  availability_zone = var.avaiable_zone_1

  tags = {
    project = "mercury"
    type = "public"
  }
}

resource "aws_subnet" "aws_subnet2" {
  vpc_id = aws_vpc.aws_vpc_mercury.id
  cidr_block = "10.0.2.0/24"
  availability_zone = var.avaiable_zone_2

  tags = {
    project = "mercury"
    type = "public"
  }
}

// Associate subnets with the route table.
resource "aws_route_table_association" "aws_route_table_association_subnet1" {
  subnet_id = aws_subnet.aws_subnet1.id
  route_table_id = aws_route_table.aws_route_table_mercury.id
}

resource "aws_route_table_association" "aws_route_table_association_subnet2" {
  subnet_id = aws_subnet.aws_subnet2.id
  route_table_id = aws_route_table.aws_route_table_mercury.id
}

// Create a defualt security group for the VPC which allows ssh connectiong and
// all outbound traffic.
resource "aws_security_group" "aws_sg_mercury_default" {
  name = "aws_sg_mercury"
  vpc_id = aws_vpc.aws_vpc_mercury.id
  description = "Security group for Mercury VPC"

  ingress {
    description = "Allow SSH connection "
    from_port = 22
    to_port = 22
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