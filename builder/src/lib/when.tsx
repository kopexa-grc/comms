/**
 * Copyright (c) Kopexa GmbH
 * SPDX-License-Identifier: BUSL-1.1
 */

type WhenProps = {
    children?: React.ReactNode
}

type ShowProps = {
    when: boolean | string | number | null | undefined
    children?: React.ReactNode
}

const isNullish = (value?: React.ReactNode): boolean => {
    if (value === null || value === undefined) return true;
    if (typeof value === 'string' && value.trim() === '') return true;
    return false;
}

export const When = (props: WhenProps) => {
    const { children } = props;

    if (isNullish(children)) {
        return null;
    }

    return children;
}

export const Show = (props: ShowProps) => {
    const { when, children } = props;

    if (isNullish(when) || when === false) {
        return null;
    }

    return children
}