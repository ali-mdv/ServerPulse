import { describe, it, expect } from "vitest";
import { mount } from "@vue/test-utils";
import Input from "@/components/ui/Input.vue";

describe("Input", () => {
  it("emits update:modelValue on input", async () => {
    const w = mount(Input, { props: { modelValue: "" } });
    await w.find("input").setValue("hello");
    const events = w.emitted("update:modelValue");
    expect(events).toBeTruthy();
    expect(events![0]).toEqual(["hello"]);
  });

  it("reflects modelValue as the input value", () => {
    const w = mount(Input, { props: { modelValue: "abc" } });
    expect((w.find("input").element as HTMLInputElement).value).toBe("abc");
  });

  it("sets aria-invalid when invalid", () => {
    const w = mount(Input, { props: { invalid: true } });
    expect(w.find("input").attributes("aria-invalid")).toBe("true");
  });

  it("omits aria-invalid when valid", () => {
    const w = mount(Input, { props: { invalid: false } });
    expect(w.find("input").attributes("aria-invalid")).toBeUndefined();
  });

  it("merges an extra class via $attrs", () => {
    const w = mount(Input, { attrs: { class: "extra-class" } });
    expect(w.find("input").classes()).toContain("extra-class");
    expect(w.find("input").classes()).toContain("input");
  });
});