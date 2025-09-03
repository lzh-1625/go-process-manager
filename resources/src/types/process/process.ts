export interface ProcessItem {
    name:         string;
    uuid:         number;
    startTime:    Date;
    user:         string;
    usage:        Usage;
    state:        State;
    termType:     TermType;
    cgroupEnable: boolean;
    memoryLimit:  number | null;
    cpuLimit:     number | null;
}

export interface State {
    state: number;
    info:  Info;
}

export enum Info {
    Empty = "",
    重启次数异常 = "重启次数异常",
}

export enum TermType {
    Pty = "pty",
}

export interface Usage {
    cpuCapacity: number;
    memCapacity: number;
    cpu:         number[] | null;
    mem:         number[] | null;
    time:        string[] | null;
}
