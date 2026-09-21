import { describe, it, expect } from "vitest";
import { mount } from "@vue/test-utils";
import PasswordInput from "@/components/ui/PasswordInput.vue";

describe("PasswordInput", () => {
  it("starts with type=password", () => {
    const w = mount(PasswordInput, { props: { modelValue: "secret" } });
    expect(w.find("input").attributes("type")).toBe("password");
  });

  it("emits update:modelValue on input", async () => {
    const w = mount(PasswordInput, { props: { modelValue: "" } });
    await w.find("input").setValue("newpw");
    const events = w.emitted("update:modelValue");
    expect(events![0]).toEqual(["newpw"]);
  });

  it("toggles type on click", async () => {
    const w = mount(PasswordInput, { props: { modelValue: "" } });
    await w.find("button").trigger("click");
    expect(w.find("input").attributes("type")).toBe("text");
    await w.find("button").trigger("click");
    expect(w.find("input").attributes("type")).toBe("password");
  });

  it("button reflects pressed state in aria-pressed", async () => {
    const w = mount(PasswordInput, { props: { modelValue: "" } });
    const btn = w.find("button");
    expect(btn.attributes("aria-pressed")).toBe("false");
    await btn.trigger("click");
    expect(btn.attributes("aria-pressed")).toBe("true");
  });

  it("sets aria-invalid when invalid", () => {
    const w = mount(PasswordInput, { props: { invalid: true } });
    expect(w.find("input").attributes("aria-invalid")).toBe("true");
  });
});