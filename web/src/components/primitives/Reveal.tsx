'use client';

import { motion, HTMLMotionProps, Variants } from 'motion/react';
import React from 'react';

const EASE = [0.22, 1, 0.36, 1] as const;

export const revealItem: Variants = {
  hidden: { opacity: 0, y: 14 },
  visible: {
    opacity: 1,
    y: 0,
    transition: { duration: 0.5, ease: EASE },
  },
};

export const revealItemSubtle: Variants = {
  hidden: { opacity: 0, y: 8 },
  visible: {
    opacity: 1,
    y: 0,
    transition: { duration: 0.4, ease: EASE },
  },
};

interface RevealStaggerProps extends HTMLMotionProps<'div'> {
  stagger?: number;
  delay?: number;
  trigger?: boolean;
  children: React.ReactNode;
}

export function RevealStagger({
  stagger = 0.07,
  delay = 0,
  trigger = true,
  children,
  ...rest
}: RevealStaggerProps) {
  const variants: Variants = {
    hidden: {},
    visible: {
      transition: {
        staggerChildren: stagger,
        delayChildren: delay,
      },
    },
  };

  return (
    <motion.div
      initial="hidden"
      animate={trigger ? 'visible' : 'hidden'}
      variants={variants}
      {...rest}
    >
      {children}
    </motion.div>
  );
}

interface RevealItemProps extends HTMLMotionProps<'div'> {
  subtle?: boolean;
  children: React.ReactNode;
}

export function RevealItem({ subtle = false, children, ...rest }: RevealItemProps) {
  return (
    <motion.div variants={subtle ? revealItemSubtle : revealItem} {...rest}>
      {children}
    </motion.div>
  );
}

interface PageEnterProps extends HTMLMotionProps<'div'> {
  children: React.ReactNode;
}

export function PageEnter({ children, ...rest }: PageEnterProps) {
  return (
    <motion.div
      initial={{ opacity: 0, y: 12 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.5, ease: EASE }}
      {...rest}
    >
      {children}
    </motion.div>
  );
}
