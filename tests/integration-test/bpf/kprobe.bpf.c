// Some kprobe test code
#include <linux/bpf.h>
#include <bpf/bpf_helpers.h>

volatile const __u8 GLOBAL_u8 = 0;
volatile const __u32 GLOBAL_u32 = 0;

SEC("kprobe/my_kprobe")
int my_kprobe(struct pt_regs *ctx) {
	bpf_printk(" KP: GLOBAL_u8: 0x%02X, GLOBAL_u32: 0x%08X", GLOBAL_u8, GLOBAL_u32);
  return 0;
}

SEC("kretprobe/my_kretprobe")
int my_kretprobe(struct pt_regs *ctx) {
	bpf_printk("KRP: GLOBAL_u8: 0x%02X, GLOBAL_u32: 0x%08X", GLOBAL_u8, GLOBAL_u32);
  return 0;
}

char _license[] SEC("license") = "GPL";
