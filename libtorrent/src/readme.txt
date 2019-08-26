To build
gcc -I D:\lt\libtorrent-1.2.1\include -I D:/lt/boost_1_70_0 -std=c++17 -O2 -c -m64 lt_wrapper.cpp
ar q liblt_wrapper.a lt_wrapper.o