int GetArchCode() {
#if defined(__x86_64__) || defined(_M_X64)
  return 1; // x86_64
#elif defined(__i386) || defined(_M_IX86)
  return 2; // x86
#elif defined(__arm__) || defined(_M_ARM)
  return 3; // ARM
#elif defined(__aarch64__) || defined(_M_ARM64)
  return 4; // ARM64
#else
  return 0; // Unknown Architecture
#endif
}
