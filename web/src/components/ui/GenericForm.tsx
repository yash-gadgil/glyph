"use client";

import { ZodObject, ZodRawShape, z } from "zod";
import { useForm, SubmitHandler, Controller } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { Input } from "@/components/primitives/Input";
import { Label } from "@/components/primitives/Label";
import { cn } from "@/lib/utils";
import GlassButton from "./GlassButton";
import GlassToggle from "./GlassToggle";


export type FieldOption = {
  label: string;
  value: any;
}

interface SelectInfo {
  type: "select",
  options: FieldOption[];
  placeholder?: string;
}

interface ToggleInfo {
  type: "toggle";
  options: [FieldOption, FieldOption];
}

interface InputInfo {
  type: "text" | "number" | "email" | "password";
  placeholder?: string;
}

type FieldInfo = SelectInfo | ToggleInfo | InputInfo;

type FieldConfig<T> = {
  name: keyof T;
  label: string;
  info: FieldInfo;
}

interface GenericFormProps<S extends ZodRawShape> {
  schema: ZodObject<S>;
  fields: FieldConfig<z.infer<ZodObject<S>>>[];
  defaultValues?: Partial<z.infer<ZodObject<S>>>;
  onSubmit: (data: z.infer<ZodObject<S>>) => void | Promise<void>;
  submitLabel?: string;
  className?: string;
  serverError?: string;
  successMessage?: string;
  fieldErrors?: Record<string, string>;
}

export function GenericForm<S extends ZodRawShape>({
  schema,
  fields,
  defaultValues,
  onSubmit,
  submitLabel = "Submit",
  className,
  serverError,
  successMessage,
  fieldErrors,
}: GenericFormProps<S>) {
  type FormData = z.infer<ZodObject<S>>;

  const {
    register,
    handleSubmit,
    control,
    formState: { errors, isSubmitting },
  } = useForm<FormData>({
    resolver: zodResolver(schema) as any,
    defaultValues: defaultValues as any,
  });
  const internalSubmit: SubmitHandler<FormData> = async (data) => {
    await onSubmit(data);
  };
  return (
    <form
      onSubmit={handleSubmit(internalSubmit)}
      noValidate
      className={cn("space-y-6 flex flex-col items-start text-base w-full lg:w-sm max-w-full", className)}
    >
      {fields.map((field) => {
        const fieldName = String(field.name);
        const error = (errors as any)[fieldName]?.message as string | undefined;
        const serverFieldError = fieldErrors?.[fieldName];
        const displayError = error || serverFieldError;
        const info = field.info;
        const inputType = info.type;

        return (
          <div key={fieldName} className="space-y-2 w-full min-w-0">
            <Label htmlFor={fieldName}>
              {field.label || fieldName.charAt(0).toUpperCase() + fieldName.slice(1)}
            </Label>

            {inputType === "select" ? (
              <select
                id={fieldName}
                {...register(fieldName as any, {
                  valueAsNumber: info.options?.some((opt) => typeof opt.value === 'number'),
                })}
                className={cn(
                  "flex h-10 w-full rounded-md border border-neutral-300 bg-white px-3 py-2 text-sm text-black transition-colors pointer-events-auto",
                  "focus:outline-none focus:ring-2 focus:ring-neutral-400 focus:ring-offset-2",
                  "disabled:cursor-not-allowed disabled:opacity-50",
                  "dark:border-neutral-700 dark:bg-zinc-800 dark:text-white dark:focus:ring-neutral-600",
                  displayError && "border-red-500 dark:border-red-500"
                )}
              >
                {info.options.map((opt) => (
                  <option key={opt.value} value={opt.value}>
                    {opt.label}
                  </option>
                ))}
              </select>
            ) : inputType === "toggle" ? (
              <Controller
                control={control}
                name={fieldName as any}
                render={({ field: { value, onChange } }) => (
                  <GlassToggle
                    value={value as boolean}
                    onChange={onChange}
                    options={info.options}
                  />
                )}
              />
            ) : (
              <Input
                id={fieldName}
                type={inputType}
                placeholder={info.placeholder}
                {...register(fieldName as any, {
                  valueAsNumber: inputType === "number",
                })}
              />
            )}

            {displayError && (
              <p role="alert" className="text-sm text-red-500 dark:text-red-400 wrap-break-word w-full max-w-full">
                {displayError}
              </p>
            )}
          </div>
        );
      })}

      {serverError && (
        <div className="w-full">
          <p role="alert" className="text-sm text-red-500 dark:text-red-400 text-center">
            {serverError}
          </p>
        </div>
      )}

      {successMessage && (
        <div className="w-full">
          <p role="status" className="text-sm text-center">
            {successMessage}
          </p>
        </div>
      )}

      <div className="w-full flex justify-center">
        <GlassButton
          type="submit"
          disabled={isSubmitting}
          text={isSubmitting ? "Submitting…" : submitLabel}
        />
      </div>
    </form>
  );
}