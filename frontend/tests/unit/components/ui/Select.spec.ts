import { describe, it, expect } from "vitest";
import { mount } from "@vue/test-utils";
import Select from "@/components/ui/Select.vue";

describe("Select", () => {
  it("emits update:modelValue on change", async () => {
    const w = mount(Select, {
      props: { modelValue: "a" },
      slots: { default: "<option value='a'>A</option><option value='b'>B</option>" },
    });
    await w.find("select").setValue("b");
    const events = w.emitted("update:modelValue");
    expect(events).toBeTruthy();
    expect(events![0]).toEqual(["b"]);
  });

  it("reflects modelValue as the select value", () => {
    const w = mount(Select, {
      props: { modelValue: "a" },
      slots: { default: "<option value='a'>A</option><option value='b'>B</option>" },
    });
    expect((w.find("select").element as HTMLSelectElement).value).toBe("a");
  });

  it("has the input base class", () => {
    const w = mount(Select, { props: { modelValue: "a" } });
    expect(w.find("select").classes()).toContain("input");
  });
});