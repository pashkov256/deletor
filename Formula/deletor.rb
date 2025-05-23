class Deletor < Formula
    desc "A powerful file deletion tool with TUI and CLI interfaces"
    homepage "https://github.com/pashkov256/deletor"
    url "https://github.com/pashkov256/deletor/archive/v1.0.0.tar.gz"
    sha256 "auto"
    license "MIT"
  
    depends_on "go" => :build
  
    def install
      system "go", "build", *std_go_args
    end
  
    test do
      system "#{bin}/deletor", "--version"
    end
  end